package g2p_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/NK8007/gofonix/pkg/g2p"
	"github.com/NK8007/gofonix/pkg/phoneme"
)

// Conformance tests pin the cross-cutting v0.1 contracts that the ADRs require
// to hold regardless of internal refactors. They complement the golden corpus:
// goldens freeze concrete outputs, these freeze structural/behavioral
// invariants.

// TestConformanceNoARPAbetLeakage asserts no public g2p type exposes RAW
// ARPAbet. Pronunciations are carried as []phoneme.Phoneme (neutral IDs), and
// no public struct may declare a field whose name suggests an ARPAbet carrier
// (Principle 3, ADR-0002, ADR-0008).
func TestConformanceNoARPAbetLeakage(t *testing.T) {
	prt := reflect.TypeOf(g2p.Pronunciation{})
	pf, ok := prt.FieldByName("Phonemes")
	if !ok {
		t.Fatal("Pronunciation has no Phonemes field")
	}
	if pf.Type != reflect.TypeOf([]phoneme.Phoneme{}) {
		t.Errorf("Pronunciation.Phonemes type = %s, want []phoneme.Phoneme", pf.Type)
	}

	forbidden := []string{"arpabet", "arpa", "symbols", "rawphones", "rawphonemes", "cmudict"}
	types := []reflect.Type{
		reflect.TypeOf(g2p.Result{}),
		reflect.TypeOf(g2p.TokenResult{}),
		reflect.TypeOf(g2p.Pronunciation{}),
		reflect.TypeOf(g2p.ResultMetadata{}),
	}
	for _, typ := range types {
		for i := 0; i < typ.NumField(); i++ {
			name := strings.ToLower(typ.Field(i).Name)
			for _, bad := range forbidden {
				if name == bad {
					t.Errorf("%s exposes forbidden ARPAbet-ish field %q", typ.Name(), typ.Field(i).Name)
				}
			}
		}
	}

	// No phoneme in any Process output carries a non-empty IPA via a field that
	// would let ARPAbet round-trip out — IPA is debug-only and we assert the
	// emitted phonemes only expose an integer ID + (debug) IPA string.
	pht := reflect.TypeOf(phoneme.Phoneme{})
	if f, ok := pht.FieldByName("ID"); !ok || f.Type.Kind() != reflect.Int {
		t.Errorf("phoneme.Phoneme.ID must be an int identity")
	}
}

// TestConformanceFeatureMaskNotUint64Alias asserts FeatureMask is an opaque
// struct, NOT a uint64 (or any integer) alias (Principle 5, ADR-0002). A
// numeric alias would let callers do arithmetic/bit ops on a mask directly and
// would break the opaque contract.
func TestConformanceFeatureMaskNotUint64Alias(t *testing.T) {
	mt := reflect.TypeOf(phoneme.FeatureMask{})
	if mt.Kind() != reflect.Struct {
		t.Fatalf("FeatureMask Kind = %s, want struct (must not be a uint64 alias)", mt.Kind())
	}
	// Defensively reject any integer kind, in case the type ever changes shape.
	switch mt.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		t.Fatalf("FeatureMask must not be an integer alias, got %s", mt.Kind())
	}
	// The opaque mask must not be directly assignable from a uint64.
	if reflect.TypeOf(uint64(0)).AssignableTo(mt) {
		t.Error("uint64 is assignable to FeatureMask; mask is not opaque")
	}
	// Its exported accessors must exist and return uint64.
	mv := reflect.ValueOf(phoneme.FromLo(1028))
	for _, name := range []string{"Lo", "Hi"} {
		m := mv.MethodByName(name)
		if !m.IsValid() {
			t.Errorf("FeatureMask is missing accessor %s()", name)
		}
	}
}

// TestConformanceCausalAllZero asserts ModeCausal yields an all-zero
// FeatureStream and every token is SourceUnknown with empty phonemes/alignment
// (ADR-0003 — the causal scaffold performs no lookup/projection in v0.1).
func TestConformanceCausalAllZero(t *testing.T) {
	e, err := g2p.New(g2p.Options{Mode: g2p.ModeCausal})
	if err != nil {
		t.Fatal(err)
	}
	for _, in := range []string{"cat", "the dog runs", "well-known 42", "caf\u00e9"} {
		res, err := e.Process(in)
		if err != nil {
			t.Fatalf("Process(%q): %v", in, err)
		}
		for i, m := range res.FeatureStream {
			if !m.IsZero() {
				t.Errorf("causal Process(%q): FeatureStream[%d] non-zero", in, i)
			}
		}
		for i, tr := range res.Tokens {
			if tr.Pronunciation.Source != g2p.SourceUnknown {
				t.Errorf("causal Process(%q): token %d source = %d, want SourceUnknown", in, i, tr.Pronunciation.Source)
			}
			if len(tr.Pronunciation.Phonemes) != 0 || len(tr.Pronunciation.Alignment) != 0 {
				t.Errorf("causal Process(%q): token %d carries phonemes/alignment, want none", in, i)
			}
		}
	}
}

// TestConformanceUnsupportedLanguage asserts a non-"en" language yields
// ErrUnsupportedLanguage (ADR-0001).
func TestConformanceUnsupportedLanguage(t *testing.T) {
	for _, lang := range []string{"pl", "fr", "EN", "en-US", "zz"} {
		_, err := g2p.New(g2p.Options{Language: lang})
		if !errors.Is(err, g2p.ErrUnsupportedLanguage) {
			t.Errorf("New(Language=%q) err = %v, want ErrUnsupportedLanguage", lang, err)
		}
	}
	// Empty language defaults to "en" and must succeed.
	if _, err := g2p.New(g2p.Options{Language: ""}); err != nil {
		t.Errorf("New(Language=\"\") err = %v, want nil (defaults to en)", err)
	}
}

// TestConformanceInvalidMode asserts an out-of-range Mode yields ErrInvalidMode
// (ADR-0001, ADR-0003).
func TestConformanceInvalidMode(t *testing.T) {
	for _, m := range []g2p.Mode{g2p.Mode(-1), g2p.Mode(3), g2p.Mode(99)} {
		_, err := g2p.New(g2p.Options{Mode: m})
		if !errors.Is(err, g2p.ErrInvalidMode) {
			t.Errorf("New(Mode=%d) err = %v, want ErrInvalidMode", m, err)
		}
	}
	// The three valid modes must construct cleanly.
	for _, m := range []g2p.Mode{g2p.ModeBatch, g2p.ModeOracle, g2p.ModeCausal} {
		if _, err := g2p.New(g2p.Options{Mode: m}); err != nil {
			t.Errorf("New(Mode=%d) err = %v, want nil", m, err)
		}
	}
}

// TestConformanceInvalidUTF8IsUnknownNotError asserts invalid UTF-8 is
// tokenized (KindUnknown bytes), never rejected: Process returns no error and
// never ErrInvalidInput in v0.1 (ADR-0006).
func TestConformanceInvalidUTF8IsUnknownNotError(t *testing.T) {
	e, err := g2p.New(g2p.Options{})
	if err != nil {
		t.Fatal(err)
	}
	inputs := []string{"\xff", "\xff\xfe", "a\xffb", "\x80\x80", "caf\xe9"}
	for _, in := range inputs {
		res, err := e.Process(in)
		if err != nil {
			t.Errorf("Process(%q) err = %v, want nil (invalid UTF-8 must not error)", in, err)
			continue
		}
		if errors.Is(err, g2p.ErrInvalidInput) {
			t.Errorf("Process(%q) returned ErrInvalidInput; v0.1 must not reject invalid UTF-8", in)
		}
		if len(res.FeatureStream) != len(in) {
			t.Errorf("Process(%q): FeatureStream len = %d, want %d", in, len(res.FeatureStream), len(in))
		}
		// At least one invalid byte must surface as KindUnknown.
		sawUnknown := false
		for _, tr := range res.Tokens {
			if tr.Kind == g2p.KindUnknown {
				sawUnknown = true
			}
		}
		if !sawUnknown {
			t.Errorf("Process(%q): expected at least one KindUnknown token for invalid UTF-8", in)
		}
	}
}

// TestConformanceBoundaryMaskNeverEmitted asserts the boundary phoneme mask
// (ID 0 -> lo=4194304) never appears in a projected FeatureStream; non-speech
// bytes get the zero mask instead (ADR-0007, ADR-0008). It also pins the
// boundary mask value via the public FeatureForPhoneme accessor.
func TestConformanceBoundaryMaskNeverEmitted(t *testing.T) {
	boundary := g2p.FeatureForPhoneme(phoneme.Phoneme{ID: 0})
	if boundary.Lo() != 4194304 || boundary.Hi() != 0 {
		t.Errorf("boundary mask = (lo=%d,hi=%d), want (lo=4194304,hi=0)", boundary.Lo(), boundary.Hi())
	}
	for _, m := range []g2p.Mode{g2p.ModeBatch, g2p.ModeOracle} {
		e, _ := g2p.New(g2p.Options{Mode: m})
		res, err := e.Process("cat. the dog! hello-world 42 it's")
		if err != nil {
			t.Fatal(err)
		}
		for i, fm := range res.FeatureStream {
			if fm.Equal(boundary) {
				t.Errorf("mode %d: FeatureStream[%d] is the boundary mask; must never be emitted", m, i)
			}
		}
	}
}

// TestConformanceFeatureStreamByteLength pins the ADR-0007 invariant that
// FeatureStream has exactly len(input) (byte, not rune) entries across modes
// and encodings, including empty and multi-byte/invalid-UTF-8 inputs.
func TestConformanceFeatureStreamByteLength(t *testing.T) {
	inputs := []string{"", "cat", "caf\u00e9", "\u2018q\u2019", "\xff\xffx", "the dog"}
	for _, m := range []g2p.Mode{g2p.ModeBatch, g2p.ModeOracle, g2p.ModeCausal} {
		e, _ := g2p.New(g2p.Options{Mode: m})
		for _, in := range inputs {
			res, err := e.Process(in)
			if err != nil {
				t.Fatalf("mode %d Process(%q): %v", m, in, err)
			}
			if len(res.FeatureStream) != len(in) {
				t.Errorf("mode %d Process(%q): FeatureStream len = %d, want %d (bytes)",
					m, in, len(res.FeatureStream), len(in))
			}
		}
	}
}

// TestConformanceMetadataStable pins the reproducibility metadata that every
// Result must carry (ADR-0001, ADR-0004, ADR-0008).
func TestConformanceMetadataStable(t *testing.T) {
	e, _ := g2p.New(g2p.Options{})
	res, err := e.Process("cat")
	if err != nil {
		t.Fatal(err)
	}
	if res.SchemaVersion != "gofonix-result-v0.1" {
		t.Errorf("SchemaVersion = %q, want gofonix-result-v0.1", res.SchemaVersion)
	}
	md := res.Metadata
	checks := map[string]string{
		"GofonixVersion":       md.GofonixVersion,
		"FeatureSchemaVersion": md.FeatureSchemaVersion,
		"DictionaryID":         md.DictionaryID,
		"DictionaryChecksum":   md.DictionaryChecksum,
		"NormalizerVersion":    md.NormalizerVersion,
		"TokenizerVersion":     md.TokenizerVersion,
		"OOVPolicy":            md.OOVPolicy,
	}
	for name, v := range checks {
		if v == "" {
			t.Errorf("Metadata.%s is empty, want non-empty", name)
		}
	}
	const wantSum = "05ca6f91559f739e6d226a57f29d84ead2721820b51c85629662bdd870a0e5b1"
	if md.DictionaryChecksum != wantSum {
		t.Errorf("DictionaryChecksum = %q, want %q", md.DictionaryChecksum, wantSum)
	}
	if md.GofonixVersion != "v0.1.0" {
		t.Errorf("GofonixVersion = %q, want v0.1.0", md.GofonixVersion)
	}
}
