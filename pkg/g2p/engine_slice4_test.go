package g2p

import (
	"reflect"
	"testing"

	"github.com/gofonix/gofonix/pkg/phoneme"
)

// boundaryMask is the boundary phoneme (ID 0) FeatureMask per ADR-0008
// (FeatBoundary, lo=4194304). It must NEVER appear in the FeatureStream in
// v0.1; non-speech bytes get the zero mask instead.
var boundaryMask = phoneme.FromLo(4194304)

// findToken returns the first token with the given surface, or fails.
func findToken(t *testing.T, res Result, surface string) TokenResult {
	t.Helper()
	for _, tr := range res.Tokens {
		if tr.Token == surface {
			return tr
		}
	}
	t.Fatalf("token %q not found in %v", surface, res.Tokens)
	return TokenResult{}
}

// assertPartitionsToken verifies a dict-hit word token's alignment is a
// contiguous, non-overlapping partition of its byte span with one span per
// phoneme and all spans within input bounds.
func assertPartitionsToken(t *testing.T, in string, tr TokenResult) {
	t.Helper()
	pr := tr.Pronunciation
	if len(pr.Alignment) != len(pr.Phonemes) {
		t.Errorf("token %q: %d alignment spans, want %d (==#phonemes)",
			tr.Token, len(pr.Alignment), len(pr.Phonemes))
		return
	}
	if len(pr.Alignment) == 0 {
		return
	}
	if pr.Alignment[0].Start != tr.Span.Start {
		t.Errorf("token %q: alignment starts at %d, want %d", tr.Token, pr.Alignment[0].Start, tr.Span.Start)
	}
	last := pr.Alignment[len(pr.Alignment)-1]
	if last.End != tr.Span.End {
		t.Errorf("token %q: alignment ends at %d, want %d", tr.Token, last.End, tr.Span.End)
	}
	for i, s := range pr.Alignment {
		if s.Start < 0 || s.End > len(in) || s.Start > s.End {
			t.Errorf("token %q: span %d = %v out of bounds [0,%d]", tr.Token, i, s, len(in))
		}
		if i > 0 && pr.Alignment[i-1].End != s.Start {
			t.Errorf("token %q: span %d not contiguous with previous", tr.Token, i)
		}
	}
}

// TestAlignmentCatThreePhonemes: "cat" with 3 phonemes and 3 bytes gives the
// alignment [0,1) [1,2) [2,3).
func TestAlignmentCatThreePhonemes(t *testing.T) {
	e, _ := New(Options{})
	res, err := e.Process("cat")
	if err != nil {
		t.Fatal(err)
	}
	tr := res.Tokens[0]
	want := []ByteSpan{{0, 1}, {1, 2}, {2, 3}}
	if !reflect.DeepEqual(tr.Pronunciation.Alignment, want) {
		t.Errorf("alignment = %v, want %v", tr.Pronunciation.Alignment, want)
	}
}

// TestAlignmentRemainderToLeading: "hello" has 5 bytes and HELLO -> HH AH L OW
// = 4 phonemes. M=5, N=4 -> base=1, R=1 -> widths 2,1,1,1; remainder to the
// leading phoneme.
func TestAlignmentRemainderToLeading(t *testing.T) {
	e, _ := New(Options{})
	res, err := e.Process("hello")
	if err != nil {
		t.Fatal(err)
	}
	tr := res.Tokens[0]
	if len(tr.Pronunciation.Phonemes) != 4 {
		t.Fatalf("hello phonemes = %d, want 4", len(tr.Pronunciation.Phonemes))
	}
	want := []ByteSpan{{0, 2}, {2, 3}, {3, 4}, {4, 5}}
	if !reflect.DeepEqual(tr.Pronunciation.Alignment, want) {
		t.Errorf("alignment = %v, want %v", tr.Pronunciation.Alignment, want)
	}
	assertPartitionsToken(t, "hello", tr)
}

// TestAlignmentItOneToOne asserts the engine produces a valid 1:1 partition for
// the short word "it" (IH T, 2 bytes, 2 phonemes). The M<N zero-width-span case
// is exercised directly in TestUniformByteAlignmentShortM (align_test.go),
// since the mini dict has no word whose phoneme count exceeds its byte length.
func TestAlignmentItOneToOne(t *testing.T) {
	e, _ := New(Options{})
	res, err := e.Process("it")
	if err != nil {
		t.Fatal(err)
	}
	tr := res.Tokens[0]
	want := []ByteSpan{{0, 1}, {1, 2}}
	if !reflect.DeepEqual(tr.Pronunciation.Alignment, want) {
		t.Errorf("alignment = %v, want %v", tr.Pronunciation.Alignment, want)
	}
}

// TestProjectionCatDotBoundaryVsZero: "cat." projects [K, AE, T] masks for the
// three letters and a ZERO mask (not the boundary mask) for ".".
func TestProjectionCatDotBoundaryVsZero(t *testing.T) {
	e, _ := New(Options{})
	const in = "cat."
	res, err := e.Process(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.FeatureStream) != len(in) {
		t.Fatalf("FeatureStream len = %d, want %d", len(res.FeatureStream), len(in))
	}
	want := []phoneme.FeatureMask{
		phoneme.FromLo(1028),    // K
		phoneme.FromLo(2146305), // AE
		phoneme.FromLo(260),     // T
		phoneme.Zero(),          // "." non-speech -> zero mask, NOT boundary
	}
	for i, fm := range res.FeatureStream {
		if !fm.Equal(want[i]) {
			t.Errorf("FeatureStream[%d] = (lo=%d,hi=%d), want (lo=%d,hi=%d)",
				i, fm.Lo(), fm.Hi(), want[i].Lo(), want[i].Hi())
		}
	}
	// The "." byte must be the zero mask and must NOT be the boundary mask.
	dot := res.FeatureStream[3]
	if dot.Equal(boundaryMask) {
		t.Error("FeatureStream for '.' is the boundary mask; v0.1 must emit zero mask")
	}
	if !dot.IsZero() {
		t.Error("FeatureStream for '.' must be zero mask")
	}
}

// TestBoundaryMaskNeverEmitted scans the FeatureStream of a varied input and
// confirms the boundary phoneme mask (lo=4194304) never appears anywhere.
func TestBoundaryMaskNeverEmitted(t *testing.T) {
	e, _ := New(Options{})
	res, err := e.Process("cat. the dog! hello-world 42")
	if err != nil {
		t.Fatal(err)
	}
	for i, fm := range res.FeatureStream {
		if fm.Equal(boundaryMask) {
			t.Errorf("FeatureStream[%d] is the boundary mask; must never be emitted in v0.1", i)
		}
	}
}

// TestWellKnownWordPunctWord: "well-known" tokenizes as word/punct/word; both
// words project their phoneme masks and the hyphen byte stays at the zero mask.
func TestWellKnownWordPunctWord(t *testing.T) {
	e, _ := New(Options{})
	const in = "well-known"
	res, err := e.Process(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.FeatureStream) != len(in) {
		t.Fatalf("FeatureStream len = %d, want %d", len(res.FeatureStream), len(in))
	}
	// Find the hyphen punctuation token and confirm its byte is the zero mask.
	hy := findToken(t, res, "-")
	if hy.Kind != KindPunctuation {
		t.Errorf("'-' Kind = %d, want KindPunctuation", hy.Kind)
	}
	for b := hy.Span.Start; b < hy.Span.End; b++ {
		if !res.FeatureStream[b].IsZero() {
			t.Errorf("hyphen byte %d projected non-zero mask, want zero", b)
		}
	}
	// "well" and "known" are dict hits -> their bytes are non-zero.
	well := findToken(t, res, "well")
	if well.Pronunciation.Source != SourceDict {
		t.Errorf("'well' Source = %d, want SourceDict", well.Pronunciation.Source)
	}
	assertPartitionsToken(t, in, well)
	for b := well.Span.Start; b < well.Span.End; b++ {
		if res.FeatureStream[b].IsZero() {
			t.Errorf("'well' byte %d is zero mask, want projected phoneme mask", b)
		}
	}
	known := findToken(t, res, "known")
	if known.Pronunciation.Source != SourceDict {
		t.Errorf("'known' Source = %d, want SourceDict", known.Pronunciation.Source)
	}
	assertPartitionsToken(t, in, known)
}

// TestOOVEmptyAlignmentZeroMask: a fallback-DECLINED OOV token has empty
// alignment and projects only zero masks. Under ADR-0009 (v0.2) a letter-only
// OOV word like "xyzzy" is now resolved by the rule fallback, so this test uses
// a non-ASCII word ("caf\u00e9"), which the fallback declines (SourceUnknown).
func TestOOVEmptyAlignmentZeroMask(t *testing.T) {
	e, _ := New(Options{})
	const in = "caf\u00e9"
	res, err := e.Process(in)
	if err != nil {
		t.Fatal(err)
	}
	pr := res.Tokens[0].Pronunciation
	if pr.Source != SourceUnknown {
		t.Errorf("Source = %d, want SourceUnknown", pr.Source)
	}
	if len(pr.Alignment) != 0 {
		t.Errorf("Alignment len = %d, want 0 (OOV)", len(pr.Alignment))
	}
	for i, fm := range res.FeatureStream {
		if !fm.IsZero() {
			t.Errorf("OOV FeatureStream[%d] non-zero, want zero", i)
		}
	}
}

// TestNonWordTokensZeroMask confirms numbers, whitespace, punctuation and
// symbols all project zero masks.
func TestNonWordTokensZeroMask(t *testing.T) {
	e, _ := New(Options{})
	const in = "  123 $ . \t"
	res, err := e.Process(in)
	if err != nil {
		t.Fatal(err)
	}
	for i, fm := range res.FeatureStream {
		if !fm.IsZero() {
			t.Errorf("non-word FeatureStream[%d] non-zero, want zero", i)
		}
	}
}

// TestBatchAndOracleProject confirms both ModeBatch and ModeOracle perform the
// annotation projection identically.
func TestBatchAndOracleProject(t *testing.T) {
	const in = "cat"
	want := []phoneme.FeatureMask{
		phoneme.FromLo(1028),    // K
		phoneme.FromLo(2146305), // AE
		phoneme.FromLo(260),     // T
	}
	for _, m := range []Mode{ModeBatch, ModeOracle} {
		e, _ := New(Options{Language: "en", Mode: m})
		res, err := e.Process(in)
		if err != nil {
			t.Fatal(err)
		}
		for i, fm := range res.FeatureStream {
			if !fm.Equal(want[i]) {
				t.Errorf("mode %d FeatureStream[%d] = (lo=%d), want (lo=%d)",
					m, i, fm.Lo(), want[i].Lo())
			}
		}
	}
}

// TestCausalNoProjection confirms ModeCausal projects nothing: every byte is the
// zero mask, alignment-bearing pronunciations are absent (SourceUnknown), and
// the batch projector is never applied to the causal stream.
func TestCausalNoProjection(t *testing.T) {
	e, _ := New(Options{Language: "en", Mode: ModeCausal})
	res, err := e.Process("cat the dog")
	if err != nil {
		t.Fatal(err)
	}
	for i, fm := range res.FeatureStream {
		if !fm.IsZero() {
			t.Errorf("ModeCausal FeatureStream[%d] non-zero, want zero", i)
		}
	}
	for i, tr := range res.Tokens {
		if tr.Pronunciation.Source != SourceUnknown {
			t.Errorf("ModeCausal token %d (%q): Source = %d, want SourceUnknown",
				i, tr.Token, tr.Pronunciation.Source)
		}
		if len(tr.Pronunciation.Alignment) != 0 {
			t.Errorf("ModeCausal token %d (%q): %d alignment spans, want 0",
				i, tr.Token, len(tr.Pronunciation.Alignment))
		}
	}
}

// TestFeatureStreamAlwaysInputLength sweeps a range of inputs (including empty,
// multi-byte UTF-8 and invalid UTF-8) and confirms len(FeatureStream)==len(in)
// in every mode.
func TestFeatureStreamAlwaysInputLength(t *testing.T) {
	inputs := []string{"", "cat", "cat.", "well-known", "hello world",
		"caf\u00e9", "\xff\xffx", "3.14 -5", "the dog runs"}
	for _, m := range []Mode{ModeBatch, ModeOracle, ModeCausal} {
		e, _ := New(Options{Language: "en", Mode: m})
		for _, in := range inputs {
			res, err := e.Process(in)
			if err != nil {
				t.Fatalf("mode %d Process(%q) error = %v", m, in, err)
			}
			if len(res.FeatureStream) != len(in) {
				t.Errorf("mode %d Process(%q): FeatureStream len = %d, want %d",
					m, in, len(res.FeatureStream), len(in))
			}
		}
	}
}

// TestAlignmentWithinBoundsAndNoOverlap sweeps a varied input and confirms that
// for every dict-hit word the alignment is a contiguous, non-overlapping
// partition within input bounds, and that no two tokens' projected bytes
// overlap (each byte is written at most once -> the per-byte mask equals the
// owning phoneme's mask, never a blended value).
func TestAlignmentWithinBoundsAndNoOverlap(t *testing.T) {
	e, _ := New(Options{})
	const in = "cat the dog hello well-known it"
	res, err := e.Process(in)
	if err != nil {
		t.Fatal(err)
	}
	// Track which bytes are covered by any token's alignment; assert no byte is
	// claimed by two different tokens (token spans are disjoint, so alignments
	// derived from them must be disjoint too).
	owner := make([]int, len(in)) // token index + 1, 0 == unclaimed
	for ti, tr := range res.Tokens {
		pr := tr.Pronunciation
		if pr.Source != SourceDict {
			continue
		}
		assertPartitionsToken(t, in, tr)
		for _, s := range pr.Alignment {
			for b := s.Start; b < s.End; b++ {
				if owner[b] != 0 {
					t.Errorf("byte %d claimed by tokens %d and %d (overlap)", b, owner[b]-1, ti)
				}
				owner[b] = ti + 1
			}
		}
	}
}

// TestMultiByteUTF8ByteAlignment exercises byte-based (not rune-aware)
// alignment. The mini dict has no multi-byte-letter words, so we drive the
// algorithm directly with a span whose length exceeds its rune count and assert
// the spans are byte offsets that may split a code point. We also confirm at
// the engine level that the FeatureStream length is the byte length, not the
// rune count, for a multi-byte input.
func TestMultiByteUTF8ByteAlignment(t *testing.T) {
	// "café" is 5 bytes but 4 runes. FeatureStream must have 5 entries.
	e, _ := New(Options{})
	const in = "caf\u00e9" // c a f é  -> bytes: 63 61 66 C3 A9
	if len(in) != 5 {
		t.Fatalf("len(%q) = %d, want 5 bytes", in, len(in))
	}
	res, err := e.Process(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.FeatureStream) != 5 {
		t.Errorf("FeatureStream len = %d, want 5 (byte count, not rune count)", len(res.FeatureStream))
	}

	// Pure byte-based alignment that splits the 2-byte 'é': a hypothetical word
	// occupying byte span [0,5) with 5 phonemes aligns to single bytes,
	// including a boundary INSIDE the 'é' code point at offset 4. This is the
	// intended byte-based (non-rune-aware) behavior of ADR-0007.
	spans := uniformByteAlignment(0, 5, 5)
	want := []ByteSpan{{0, 1}, {1, 2}, {2, 3}, {3, 4}, {4, 5}}
	if !reflect.DeepEqual(spans, want) {
		t.Errorf("byte alignment = %v, want %v", spans, want)
	}
	// Byte offset 4 falls in the middle of the 2-byte 'é' (bytes 3 and 4); the
	// alignment cuts it without regard for rune boundaries.
	if spans[3].End != 4 || spans[4].Start != 4 {
		t.Errorf("expected a span boundary at byte 4 (mid code point), got %v", spans)
	}
}
