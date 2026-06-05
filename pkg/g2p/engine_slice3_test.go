package g2p

import (
	"testing"

	"github.com/gofonix/gofonix/internal/lang/en/fallback"
)

// Phase 2 / Slice 3: deterministic OOV rule-fallback integration into Process()
// (ADR-0009). These tests pin metadata, provenance (SourceDict wins over
// fallback, fallback-resolved -> SourceRuleFallback, declined -> SourceUnknown),
// fallback phoneme counts, alignment/projection mechanics shared with dict hits,
// and the ModeCausal scaffold-only invariant.
//
// NOTE on API shape: the engine fixes its Mode at construction
// (New(Options{Mode: ...})) and Process takes only the input string. The task's
// pseudo-signature engine.Process(ModeBatch, "ship") is expressed here as
// New(Options{Mode: ModeBatch}).Process("ship").

func batchEngine(t *testing.T) *Engine {
	t.Helper()
	e, err := New(Options{Language: "en", Mode: ModeBatch})
	if err != nil {
		t.Fatalf("New(ModeBatch) error = %v", err)
	}
	return e
}

func process(t *testing.T, e *Engine, in string) Result {
	t.Helper()
	res, err := e.Process(in)
	if err != nil {
		t.Fatalf("Process(%q) error = %v", in, err)
	}
	return res
}

// ---------------------------------------------------------------------------
// A. Metadata
// ---------------------------------------------------------------------------

// TestSlice3MetadataBatch confirms the active-fallback metadata: the rule
// version string and the amended OOV policy (ADR-0009).
func TestSlice3MetadataBatch(t *testing.T) {
	e := batchEngine(t)
	res := process(t, e, "ship")
	if res.Metadata.FallbackRulesVersion != "fallback-en-v0.2" {
		t.Errorf("FallbackRulesVersion = %q, want %q",
			res.Metadata.FallbackRulesVersion, "fallback-en-v0.2")
	}
	if res.Metadata.FallbackRulesVersion != fallback.RulesVersion {
		t.Errorf("FallbackRulesVersion = %q, want fallback.RulesVersion %q",
			res.Metadata.FallbackRulesVersion, fallback.RulesVersion)
	}
	if res.Metadata.OOVPolicy != "rule-fallback-then-unknown" {
		t.Errorf("OOVPolicy = %q, want %q",
			res.Metadata.OOVPolicy, "rule-fallback-then-unknown")
	}
}

// TestSlice3MetadataOracle confirms ModeOracle also reports the active fallback.
func TestSlice3MetadataOracle(t *testing.T) {
	e, _ := New(Options{Language: "en", Mode: ModeOracle})
	res := process(t, e, "ship")
	if res.Metadata.FallbackRulesVersion != "fallback-en-v0.2" {
		t.Errorf("Oracle FallbackRulesVersion = %q, want %q",
			res.Metadata.FallbackRulesVersion, "fallback-en-v0.2")
	}
	if res.Metadata.OOVPolicy != "rule-fallback-then-unknown" {
		t.Errorf("Oracle OOVPolicy = %q, want %q",
			res.Metadata.OOVPolicy, "rule-fallback-then-unknown")
	}
}

// TestSlice3MetadataCausalInactive confirms ModeCausal reports the fallback as
// inactive (empty string) since it never applies the fallback (ADR-0003/0009).
func TestSlice3MetadataCausalInactive(t *testing.T) {
	e, _ := New(Options{Language: "en", Mode: ModeCausal})
	res := process(t, e, "ship")
	if res.Metadata.FallbackRulesVersion != "" {
		t.Errorf("Causal FallbackRulesVersion = %q, want empty string",
			res.Metadata.FallbackRulesVersion)
	}
}

// ---------------------------------------------------------------------------
// B. Source behavior
// ---------------------------------------------------------------------------

// TestSlice3SourceBehavior pins provenance across dict hits, fallback-eligible
// OOV words, and declined tokens (ADR-0009 Provenance / Failure Cases).
func TestSlice3SourceBehavior(t *testing.T) {
	e := batchEngine(t)
	cases := []struct {
		in      string // input string
		surface string // token surface to inspect
		want    Source
	}{
		{"cat", "cat", SourceDict},                // dict hit
		{"the", "the", SourceDict},                // dict hit (mini dict)
		{"ship", "ship", SourceRuleFallback},      // OOV eligible -> fallback
		{"MP3", "MP3", SourceUnknown},             // mixed alphanumeric -> declined
		{"H2O", "H2O", SourceUnknown},             // mixed alphanumeric -> declined
		{"caf\u00e9", "caf\u00e9", SourceUnknown}, // non-ASCII -> declined
	}
	for _, c := range cases {
		res := process(t, e, c.in)
		tr := findToken(t, res, c.surface)
		if tr.Pronunciation.Source != c.want {
			t.Errorf("input %q token %q: Source = %d, want %d",
				c.in, c.surface, tr.Pronunciation.Source, c.want)
		}
	}
}

// TestSlice3PunctuationWhitespaceUnknown confirms punctuation and whitespace
// tokens never trigger the fallback and stay SourceUnknown.
func TestSlice3PunctuationWhitespaceUnknown(t *testing.T) {
	e := batchEngine(t)
	res := process(t, e, ". \t")
	for _, tr := range res.Tokens {
		if tr.Kind == KindPunctuation || tr.Kind == KindWhitespace {
			if tr.Pronunciation.Source != SourceUnknown {
				t.Errorf("token %q (kind %d): Source = %d, want SourceUnknown",
					tr.Token, tr.Kind, tr.Pronunciation.Source)
			}
			if len(tr.Pronunciation.Phonemes) != 0 {
				t.Errorf("token %q: %d phonemes, want 0",
					tr.Token, len(tr.Pronunciation.Phonemes))
			}
		}
	}
}

// ---------------------------------------------------------------------------
// C. Fallback pronunciation (phoneme counts + non-zero masks)
// ---------------------------------------------------------------------------

// TestSlice3FallbackPhonemeCounts pins the phoneme counts for representative
// fallback words against the frozen fallback-en-v0.2 table (ADR-0009):
//
//	ship  -> SH IH P            (3)
//	thing -> DH IH NG           (3)
//	phone -> F AO N EH          (4)
//	xenon -> K S EH N AO N      (6)  (x -> K S)
func TestSlice3FallbackPhonemeCounts(t *testing.T) {
	e := batchEngine(t)
	cases := []struct {
		word string
		want int
	}{
		{"ship", 3},
		{"thing", 3},
		{"phone", 4},
		{"xenon", 6},
	}
	for _, c := range cases {
		res := process(t, e, c.word)
		tr := findToken(t, res, c.word)
		if tr.Pronunciation.Source != SourceRuleFallback {
			t.Errorf("%q: Source = %d, want SourceRuleFallback",
				c.word, tr.Pronunciation.Source)
		}
		if got := len(tr.Pronunciation.Phonemes); got != c.want {
			t.Errorf("%q: phoneme count = %d, want %d", c.word, got, c.want)
		}
	}
}

// TestSlice3FallbackNonZeroMasks confirms a fallback word projects non-zero
// FeatureMasks over its byte positions in ModeBatch (identical mechanics to
// dict hits).
func TestSlice3FallbackNonZeroMasks(t *testing.T) {
	e := batchEngine(t)
	const in = "ship"
	res := process(t, e, in)
	if len(res.FeatureStream) != len(in) {
		t.Fatalf("FeatureStream len = %d, want %d", len(res.FeatureStream), len(in))
	}
	tr := findToken(t, res, in)
	for b := tr.Span.Start; b < tr.Span.End; b++ {
		if res.FeatureStream[b].IsZero() {
			t.Errorf("fallback byte %d is zero mask, want projected phoneme mask", b)
		}
	}
}

// ---------------------------------------------------------------------------
// D. Alignment / projection
// ---------------------------------------------------------------------------

// TestSlice3FallbackAlignmentLength confirms a fallback word's alignment length
// equals its phoneme count and partitions its byte span exactly like a dict hit
// (ADR-0007).
func TestSlice3FallbackAlignmentLength(t *testing.T) {
	e := batchEngine(t)
	for _, word := range []string{"ship", "thing", "phone", "xenon"} {
		res := process(t, e, word)
		tr := findToken(t, res, word)
		pr := tr.Pronunciation
		if len(pr.Alignment) != len(pr.Phonemes) {
			t.Errorf("%q: alignment len = %d, want %d (==#phonemes)",
				word, len(pr.Alignment), len(pr.Phonemes))
		}
		// Reuse the slice4 partition checker: contiguous, non-overlapping,
		// in-bounds, spanning the whole token byte range.
		assertPartitionsToken(t, word, tr)
	}
}

// TestSlice3FallbackThenZeroMask confirms "ship." resolves the fallback on
// "ship" and emits a zero mask on the trailing "." (ADR-0009 Output Semantics).
func TestSlice3FallbackThenZeroMask(t *testing.T) {
	e := batchEngine(t)
	const in = "ship."
	res := process(t, e, in)
	if len(res.FeatureStream) != len(in) {
		t.Fatalf("FeatureStream len = %d, want %d", len(res.FeatureStream), len(in))
	}
	ship := findToken(t, res, "ship")
	if ship.Pronunciation.Source != SourceRuleFallback {
		t.Errorf("'ship' Source = %d, want SourceRuleFallback", ship.Pronunciation.Source)
	}
	for b := ship.Span.Start; b < ship.Span.End; b++ {
		if res.FeatureStream[b].IsZero() {
			t.Errorf("'ship' byte %d is zero mask, want projected phoneme mask", b)
		}
	}
	dot := findToken(t, res, ".")
	if dot.Kind != KindPunctuation {
		t.Errorf("'.' Kind = %d, want KindPunctuation", dot.Kind)
	}
	for b := dot.Span.Start; b < dot.Span.End; b++ {
		if !res.FeatureStream[b].IsZero() {
			t.Errorf("'.' byte %d projected non-zero mask, want zero", b)
		}
		if res.FeatureStream[b].Equal(boundaryMask) {
			t.Errorf("'.' byte %d is boundary mask; must be zero mask", b)
		}
	}
}

// ---------------------------------------------------------------------------
// E. ModeCausal scaffold-only
// ---------------------------------------------------------------------------

// TestSlice3CausalFallbackAllZero confirms a fallback-eligible word in
// ModeCausal produces an all-zero FeatureStream and is NOT mixed with the
// annotation fallback. Documented choice: the token stays SourceUnknown with
// empty phonemes in ModeCausal (the simplest policy; the causal scaffold never
// applies the fallback, and the annotation stream is never mixed with the
// causal scaffold) (ADR-0003/0009).
func TestSlice3CausalFallbackAllZero(t *testing.T) {
	e, _ := New(Options{Language: "en", Mode: ModeCausal})
	const in = "ship thing phone"
	res := process(t, e, in)
	for i, fm := range res.FeatureStream {
		if !fm.IsZero() {
			t.Errorf("ModeCausal FeatureStream[%d] non-zero, want all-zero", i)
		}
	}
	for _, word := range []string{"ship", "thing", "phone"} {
		tr := findToken(t, res, word)
		if tr.Pronunciation.Source != SourceUnknown {
			t.Errorf("ModeCausal %q: Source = %d, want SourceUnknown (scaffold-only)",
				word, tr.Pronunciation.Source)
		}
		if len(tr.Pronunciation.Phonemes) != 0 {
			t.Errorf("ModeCausal %q: %d phonemes, want 0 (scaffold-only)",
				word, len(tr.Pronunciation.Phonemes))
		}
		if len(tr.Pronunciation.Alignment) != 0 {
			t.Errorf("ModeCausal %q: %d alignment spans, want 0 (scaffold-only)",
				word, len(tr.Pronunciation.Alignment))
		}
	}
}
