package g2p

import (
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/NK8007/gofonix/pkg/phoneme"
)

func TestNewDefaultsAndValidation(t *testing.T) {
	// Empty language defaults to "en".
	e, err := New(Options{})
	if err != nil {
		t.Fatalf("New(Options{}) error = %v, want nil", err)
	}
	if e.language != "en" {
		t.Errorf("default language = %q, want \"en\"", e.language)
	}

	// Unsupported language.
	if _, err := New(Options{Language: "fr"}); !errors.Is(err, ErrUnsupportedLanguage) {
		t.Errorf("New(fr) error = %v, want ErrUnsupportedLanguage", err)
	}

	// Invalid mode.
	if _, err := New(Options{Language: "en", Mode: Mode(99)}); !errors.Is(err, ErrInvalidMode) {
		t.Errorf("New(mode=99) error = %v, want ErrInvalidMode", err)
	}

	// All valid modes accepted.
	for _, m := range []Mode{ModeBatch, ModeOracle, ModeCausal} {
		if _, err := New(Options{Language: "en", Mode: m}); err != nil {
			t.Errorf("New(mode=%d) error = %v, want nil", m, err)
		}
	}
}

func TestFeatureStreamLength(t *testing.T) {
	e, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}
	inputs := []string{"", "cat", "hello world", "don't", "caf\u00e9", "\xff\xffx", "3.14 -5"}
	for _, in := range inputs {
		res, err := e.Process(in)
		if err != nil {
			t.Fatalf("Process(%q) error = %v", in, err)
		}
		if len(res.FeatureStream) != len(in) {
			t.Errorf("Process(%q): FeatureStream len = %d, want %d",
				in, len(res.FeatureStream), len(in))
		}
	}
}

// TestCausalAllZeroFeatureStream pins the ModeCausal scaffold invariant: even
// with dictionary words present, the causal FeatureStream is all-zero (no
// projection; ADR-0003). ModeBatch/ModeOracle projection is asserted elsewhere.
func TestCausalAllZeroFeatureStream(t *testing.T) {
	e, err := New(Options{Language: "en", Mode: ModeCausal})
	if err != nil {
		t.Fatal(err)
	}
	res, err := e.Process("hello world cat")
	if err != nil {
		t.Fatal(err)
	}
	for i, fm := range res.FeatureStream {
		if !fm.IsZero() {
			t.Errorf("ModeCausal FeatureStream[%d] non-zero (lo=%d,hi=%d), want zero",
				i, fm.Lo(), fm.Hi())
		}
	}
}

func TestCausalModeZeroMasks(t *testing.T) {
	e, err := New(Options{Language: "en", Mode: ModeCausal})
	if err != nil {
		t.Fatal(err)
	}
	res, err := e.Process("the quick brown fox")
	if err != nil {
		t.Fatal(err)
	}
	zero := phoneme.Zero()
	for i, fm := range res.FeatureStream {
		if !fm.Equal(zero) {
			t.Errorf("ModeCausal FeatureStream[%d] != zero mask", i)
		}
	}
}

// TestTokenInvariants pins the Slice-3-invariant subset that survives the move
// from Slice 1's all-SourceUnknown world: every token's surface matches its
// byte span, every Variant is 0, and Alignment stays empty (Slice 4 owns
// alignment). Pronunciation Source now depends on dictionary membership, so it
// is NOT asserted here (see TestDictHitPronunciation / TestOOVUnknown).
func TestTokenInvariants(t *testing.T) {
	e, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}
	res, err := e.Process("cat don't well-known 123")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Tokens) == 0 {
		t.Fatal("expected tokens, got none")
	}
	for i, tr := range res.Tokens {
		// Slice 4: len(Alignment) == len(Phonemes) for every token.
		if len(tr.Pronunciation.Alignment) != len(tr.Pronunciation.Phonemes) {
			t.Errorf("token %d (%q): %d alignment spans, want %d (== #phonemes)",
				i, tr.Token, len(tr.Pronunciation.Alignment), len(tr.Pronunciation.Phonemes))
		}
		if tr.Pronunciation.Variant != 0 {
			t.Errorf("token %d (%q): Variant = %d, want 0", i, tr.Token, tr.Pronunciation.Variant)
		}
		// Surface must match the byte span exactly.
		if tr.Token != res.Input[tr.Span.Start:tr.Span.End] {
			t.Errorf("token %d: surface %q != input span [%d,%d)", i, tr.Token, tr.Span.Start, tr.Span.End)
		}
	}
}

func TestMetadata(t *testing.T) {
	e, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}
	res, _ := e.Process("cat")
	md := res.Metadata
	// SchemaVersion is a top-level field on Result, not inside Metadata.
	if res.SchemaVersion != "gofonix-result-v0.1" {
		t.Errorf("Result.SchemaVersion = %q, want %q", res.SchemaVersion, "gofonix-result-v0.1")
	}
	checks := map[string]struct{ got, want string }{
		"GofonixVersion":       {md.GofonixVersion, "v0.3.1-alpha"},
		"FeatureSchemaVersion": {md.FeatureSchemaVersion, "v0.1-en"},
		"NormalizerVersion":    {md.NormalizerVersion, "v0.1"},
		"TokenizerVersion":     {md.TokenizerVersion, "v0.1"},
		// ADR-0009 (Phase 2 / Slice 3) amends ADR-0005: the English OOV policy is
		// now rule-fallback-then-unknown and the active fallback rule-set version
		// is surfaced in metadata.
		"OOVPolicy":            {md.OOVPolicy, "rule-fallback-then-unknown"},
		"FallbackRulesVersion": {md.FallbackRulesVersion, "fallback-en-v0.2"},
		// Slice 2: the embedded mini CMUdict is now loaded and recorded.
		"DictionaryID": {md.DictionaryID, "cmudict-mini-v0.1"},
	}
	for name, c := range checks {
		if c.got != c.want {
			t.Errorf("Metadata.%s = %q, want %q", name, c.got, c.want)
		}
	}
	// DictionaryChecksum is now a non-empty, stable SHA-256 hex string.
	if len(md.DictionaryChecksum) != 64 {
		t.Errorf("Metadata.DictionaryChecksum = %q (len %d), want 64-char SHA-256 hex",
			md.DictionaryChecksum, len(md.DictionaryChecksum))
	}
}

// TestConcurrentProcess spawns 10 goroutines each calling Process; run with
// -race to detect any data race. The Engine must be safe for concurrent use.
func TestConcurrentProcess(t *testing.T) {
	e, err := New(Options{Language: "en", Mode: ModeBatch})
	if err != nil {
		t.Fatal(err)
	}
	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				res, err := e.Process("hello world")
				if err != nil {
					t.Errorf("Process error: %v", err)
					return
				}
				if len(res.FeatureStream) != len("hello world") {
					t.Errorf("FeatureStream len = %d, want %d", len(res.FeatureStream), len("hello world"))
					return
				}
			}
		}()
	}
	wg.Wait()
}

// TestEngineMetadata verifies Slice 2 dictionary provenance: Process records the
// embedded mini CMUdict ID and a non-empty, stable SHA-256 checksum (ADR-0004).
func TestEngineMetadata(t *testing.T) {
	e, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}
	res, err := e.Process("cat")
	if err != nil {
		t.Fatal(err)
	}
	if res.Metadata.DictionaryID != "cmudict-mini-v0.1" {
		t.Errorf("DictionaryID = %q, want %q", res.Metadata.DictionaryID, "cmudict-mini-v0.1")
	}
	if res.Metadata.DictionaryChecksum == "" {
		t.Error("DictionaryChecksum is empty, want non-empty SHA-256 hex")
	}
	// The checksum must be stable across Process calls.
	res2, _ := e.Process("dog")
	if res.Metadata.DictionaryChecksum != res2.Metadata.DictionaryChecksum {
		t.Errorf("DictionaryChecksum not stable: %q != %q",
			res.Metadata.DictionaryChecksum, res2.Metadata.DictionaryChecksum)
	}
}

// TestFeatureStreamLengthInvariant confirms the FeatureStream always has exactly
// len(input) entries even with a mix of dict words, OOV, punctuation, and
// numbers (ADR-0007). Slice 4 projects masks for dict words; the length
// invariant must still hold.
func TestFeatureStreamLengthInvariant(t *testing.T) {
	e, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}
	const input = "cat the don't well-known 123"
	res, err := e.Process(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.FeatureStream) != len(input) {
		t.Fatalf("FeatureStream len = %d, want %d", len(res.FeatureStream), len(input))
	}
}

// --- Slice 3 tests (ARPAbet mapping, dictionary integration) ---

// phonemeIDs is a small test helper extracting neutral IDs from a phoneme slice.
func phonemeIDs(ps []phoneme.Phoneme) []int {
	out := make([]int, len(ps))
	for i, p := range ps {
		out[i] = p.ID
	}
	return out
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestDictHitPronunciation pins the canonical dictionary-hit path: "cat" maps to
// neutral phoneme IDs [K=20, AE=2, T=31] with Source SourceDict and an empty
// Alignment (per-phoneme alignment is Slice 4).
func TestDictHitPronunciation(t *testing.T) {
	e, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}
	res, err := e.Process("cat")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Tokens) != 1 {
		t.Fatalf("got %d tokens, want 1", len(res.Tokens))
	}
	pr := res.Tokens[0].Pronunciation
	if pr.Source != SourceDict {
		t.Errorf("Source = %d, want SourceDict", pr.Source)
	}
	got := phonemeIDs(pr.Phonemes)
	want := []int{20, 2, 31}
	if !equalInts(got, want) {
		t.Errorf("phoneme IDs = %v, want %v", got, want)
	}
	// Slice 4: "cat" (3 bytes, 3 phonemes) aligns to [0,1) [1,2) [2,3).
	wantAlign := []ByteSpan{{0, 1}, {1, 2}, {2, 3}}
	if !reflect.DeepEqual(pr.Alignment, wantAlign) {
		t.Errorf("Alignment = %v, want %v", pr.Alignment, wantAlign)
	}
	if pr.Variant != 0 {
		t.Errorf("Variant = %d, want 0", pr.Variant)
	}
}

// TestDictHitFeatureStream confirms that in Slice 4 a dictionary hit projects
// each phoneme's v0.1-en FeatureMask over its alignment span: "cat" yields a
// 3-byte FeatureStream of [K, AE, T] masks.
func TestDictHitFeatureStream(t *testing.T) {
	e, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}
	res, err := e.Process("cat")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.FeatureStream) != len("cat") {
		t.Fatalf("FeatureStream len = %d, want %d", len(res.FeatureStream), len("cat"))
	}
	// Authoritative v0.1-en lo words (ADR-0008): K=1028, AE=2146305, T=260.
	want := []phoneme.FeatureMask{
		phoneme.FromLo(1028),    // K
		phoneme.FromLo(2146305), // AE
		phoneme.FromLo(260),     // T
	}
	for i, fm := range res.FeatureStream {
		if !fm.Equal(want[i]) {
			t.Errorf("FeatureStream[%d] = (lo=%d,hi=%d), want (lo=%d,hi=%d)",
				i, fm.Lo(), fm.Hi(), want[i].Lo(), want[i].Hi())
		}
	}
}

// TestOOVUnknown confirms a fallback-INELIGIBLE out-of-vocabulary token yields
// SourceUnknown with an empty phoneme sequence. Under ADR-0009 (v0.2) the
// unknown-only behaviour is retained only for tokens the rule fallback declines;
// "H2O" carries an interior digit and is therefore declined (mixed alphanumeric).
func TestOOVUnknown(t *testing.T) {
	e, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}
	res, err := e.Process("H2O")
	if err != nil {
		t.Fatal(err)
	}
	pr := res.Tokens[0].Pronunciation
	if pr.Source != SourceUnknown {
		t.Errorf("Source = %d, want SourceUnknown", pr.Source)
	}
	if len(pr.Phonemes) != 0 {
		t.Errorf("phonemes = %v, want empty", phonemeIDs(pr.Phonemes))
	}
}

// TestOOVRuleFallback confirms that under ADR-0009 (v0.2) a fallback-ELIGIBLE
// out-of-vocabulary English word is now resolved through the deterministic rule
// fallback (SourceRuleFallback) rather than declined. "xyzzy" was SourceUnknown
// under the v0.1 unknown-only policy; it is letter-only and now resolves.
func TestOOVRuleFallback(t *testing.T) {
	e, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}
	res, err := e.Process("xyzzy")
	if err != nil {
		t.Fatal(err)
	}
	pr := res.Tokens[0].Pronunciation
	if pr.Source != SourceRuleFallback {
		t.Errorf("Source = %d, want SourceRuleFallback", pr.Source)
	}
	if len(pr.Phonemes) == 0 {
		t.Error("phonemes empty, want a fallback pronunciation")
	}
}

// TestNonWordUnknown confirms a non-word token (punctuation ".") is never
// resolved against the dictionary and stays SourceUnknown with no phonemes.
func TestNonWordUnknown(t *testing.T) {
	e, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}
	res, err := e.Process(".")
	if err != nil {
		t.Fatal(err)
	}
	pr := res.Tokens[0].Pronunciation
	if res.Tokens[0].Kind != KindPunctuation {
		t.Fatalf("Kind = %d, want KindPunctuation", res.Tokens[0].Kind)
	}
	if pr.Source != SourceUnknown {
		t.Errorf("Source = %d, want SourceUnknown", pr.Source)
	}
	if len(pr.Phonemes) != 0 {
		t.Errorf("phonemes = %v, want empty", phonemeIDs(pr.Phonemes))
	}
}

// TestCausalScaffoldNoLookup confirms ModeCausal performs no dictionary lookup:
// even dictionary words are SourceUnknown, and the FeatureStream stays all-zero
// (ADR-0003 causal scaffold).
func TestCausalScaffoldNoLookup(t *testing.T) {
	e, err := New(Options{Language: "en", Mode: ModeCausal})
	if err != nil {
		t.Fatal(err)
	}
	res, err := e.Process("cat the dog")
	if err != nil {
		t.Fatal(err)
	}
	for i, tr := range res.Tokens {
		if tr.Pronunciation.Source != SourceUnknown {
			t.Errorf("token %d (%q): Source = %d, want SourceUnknown in ModeCausal", i, tr.Token, tr.Pronunciation.Source)
		}
		if len(tr.Pronunciation.Phonemes) != 0 {
			t.Errorf("token %d (%q): phonemes = %v, want empty", i, tr.Token, phonemeIDs(tr.Pronunciation.Phonemes))
		}
	}
	for i, fm := range res.FeatureStream {
		if !fm.IsZero() {
			t.Errorf("ModeCausal FeatureStream[%d] non-zero, want zero", i)
		}
	}
}

// TestOracleModeDictHit confirms ModeOracle resolves dictionary hits exactly
// like ModeBatch (both are upper-bound modes that may consult the dictionary).
func TestOracleModeDictHit(t *testing.T) {
	e, err := New(Options{Language: "en", Mode: ModeOracle})
	if err != nil {
		t.Fatal(err)
	}
	res, err := e.Process("dog")
	if err != nil {
		t.Fatal(err)
	}
	pr := res.Tokens[0].Pronunciation
	if pr.Source != SourceDict {
		t.Errorf("Source = %d, want SourceDict", pr.Source)
	}
	// DOG -> D AO1 G -> IDs 9, 4, 15
	got := phonemeIDs(pr.Phonemes)
	want := []int{9, 4, 15}
	if !equalInts(got, want) {
		t.Errorf("phoneme IDs = %v, want %v", got, want)
	}
}

// TestARPAbetNotInPublicAPI guards Principle 3: no public g2p type exposes a raw
// ARPAbet string field. We reflect over Result, TokenResult, and Pronunciation
// and require that the only string-typed fields are the documented neutral
// surfaces/versions — never a phoneme symbol carrier. The phoneme identity in
// the public API is the neutral integer phoneme.Phoneme.ID; phoneme.Phoneme.IPA
// is a debug rendering, not ARPAbet.
func TestARPAbetNotInPublicAPI(t *testing.T) {
	// Pronunciation must carry phonemes as []phoneme.Phoneme (neutral IDs),
	// never as a []string of ARPAbet symbols.
	prt := reflect.TypeOf(Pronunciation{})
	pf, ok := prt.FieldByName("Phonemes")
	if !ok {
		t.Fatal("Pronunciation has no Phonemes field")
	}
	if pf.Type != reflect.TypeOf([]phoneme.Phoneme{}) {
		t.Errorf("Pronunciation.Phonemes type = %s, want []phoneme.Phoneme", pf.Type)
	}
	// No public g2p struct may declare a field whose name suggests a raw ARPAbet
	// carrier (e.g. "ARPAbet", "Arpa", "Symbols", "RawPhones").
	forbidden := []string{"arpabet", "arpa", "symbols", "rawphones", "rawphonemes"}
	for _, typ := range []reflect.Type{
		reflect.TypeOf(Result{}),
		reflect.TypeOf(TokenResult{}),
		reflect.TypeOf(Pronunciation{}),
	} {
		for i := 0; i < typ.NumField(); i++ {
			name := strings.ToLower(typ.Field(i).Name)
			for _, bad := range forbidden {
				if name == bad {
					t.Errorf("%s exposes forbidden ARPAbet field %q", typ.Name(), typ.Field(i).Name)
				}
			}
		}
	}
}
