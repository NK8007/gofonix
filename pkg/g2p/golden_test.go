package g2p_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/NK8007/gofonix/internal/golden"
	"github.com/NK8007/gofonix/pkg/g2p"
)

// updateGolden is set by the -update flag OR the GOFONIX_UPDATE_GOLDEN=1 env
// var. When true, the golden tests REGENERATE the .jsonl fixtures from live
// Process output instead of comparing against them. The default is false: a
// normal `go test` run NEVER rewrites a golden file and FAILS on any mismatch.
var updateGolden = flag.Bool("update", false,
	"regenerate testdata/golden/*.jsonl from live Process output instead of comparing")

func goldenUpdateEnabled() bool {
	return *updateGolden || os.Getenv("GOFONIX_UPDATE_GOLDEN") == "1"
}

// goldenDir is the directory holding the .jsonl golden corpus. The g2p test
// package lives in pkg/g2p, so testdata/golden is three levels up.
func goldenDir() string { return filepath.Join("..", "..", "testdata", "golden") }

// goldenCase is one (input, mode) pair to be (de)serialized for a golden file.
type goldenCase struct {
	input string
	mode  g2p.Mode
}

// goldenSuite groups the cases that share one .jsonl file. The order is the
// line order written to / expected from the file.
type goldenSuite struct {
	file  string
	cases []goldenCase
}

// goldenSuites defines the entire golden corpus. Inputs cover dictionary words,
// atypical surface forms (curly apostrophes, hyphen runs, trailing punctuation,
// acronyms), OOV words, non-word runs, and a causal-mode mirror that must be
// all-zero (ADR-0003).
func goldenSuites() []goldenSuite {
	// The right single quotation mark U+2019 used by "don't"/"it's" curly forms.
	const rsquo = "\u2019"
	return []goldenSuite{
		{
			file: "words.jsonl",
			cases: []goldenCase{
				{"cat", g2p.ModeBatch},
				{"a", g2p.ModeBatch},
				{"the", g2p.ModeBatch},
				{"dog", g2p.ModeBatch},
				{"it", g2p.ModeBatch},
			},
		},
		{
			file: "atypical.jsonl",
			cases: []goldenCase{
				{"cat.", g2p.ModeBatch},
				{"don't", g2p.ModeBatch},            // ASCII apostrophe
				{"it" + rsquo + "s", g2p.ModeBatch}, // U+2019 right single quote
				{"'hello'", g2p.ModeBatch},
				{"well-known", g2p.ModeBatch},
				{"well--known", g2p.ModeBatch},
				{"NASA", g2p.ModeBatch},
				{"hello", g2p.ModeBatch},
			},
		},
		{
			file: "oov.jsonl",
			cases: []goldenCase{
				{"qwertyx", g2p.ModeBatch},
				{"NASA", g2p.ModeBatch},
			},
		},
		{
			file: "nonword.jsonl",
			cases: []goldenCase{
				{"123", g2p.ModeBatch},
				{"3.14", g2p.ModeBatch},
				{"3rd", g2p.ModeBatch},
				{"MP3", g2p.ModeBatch},
				{" ", g2p.ModeBatch},
				{".", g2p.ModeBatch},
			},
		},
		{
			file: "causal.jsonl",
			cases: []goldenCase{
				{"cat", g2p.ModeCausal},
				{"the dog", g2p.ModeCausal},
				{"it" + rsquo + "s", g2p.ModeCausal},
				{"well-known", g2p.ModeCausal},
				{"123", g2p.ModeCausal},
			},
		},
	}
}

// processCase runs one golden case through a fresh Engine and returns the live
// golden.Record.
func processCase(t *testing.T, c goldenCase) golden.Record {
	t.Helper()
	e, err := g2p.New(g2p.Options{Language: "en", Mode: c.mode})
	if err != nil {
		t.Fatalf("New(mode=%v): %v", c.mode, err)
	}
	res, err := e.Process(c.input)
	if err != nil {
		t.Fatalf("Process(%q): %v", c.input, err)
	}
	return golden.FromResult(res)
}

// TestGolden either regenerates (with -update / GOFONIX_UPDATE_GOLDEN=1) or
// verifies every golden file against live Process output, line by line, with an
// exact deep-equality match.
func TestGolden(t *testing.T) {
	for _, suite := range goldenSuites() {
		suite := suite
		t.Run(suite.file, func(t *testing.T) {
			path := filepath.Join(goldenDir(), suite.file)

			// Build the live records once; used for both update and compare.
			live := make([]golden.Record, len(suite.cases))
			for i, c := range suite.cases {
				live[i] = processCase(t, c)
			}

			if goldenUpdateEnabled() {
				var buf bytes.Buffer
				if err := golden.WriteAll(&buf, live); err != nil {
					t.Fatalf("WriteAll: %v", err)
				}
				if err := os.MkdirAll(goldenDir(), 0o755); err != nil {
					t.Fatalf("MkdirAll: %v", err)
				}
				if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
					t.Fatalf("WriteFile %s: %v", path, err)
				}
				t.Logf("updated %s (%d records)", path, len(live))
				return
			}

			f, err := os.Open(path)
			if err != nil {
				t.Fatalf("open golden %s: %v (run with -update to generate)", path, err)
			}
			defer f.Close()
			want, err := golden.ReadAll(f)
			if err != nil {
				t.Fatalf("read golden %s: %v", path, err)
			}

			if len(want) != len(live) {
				t.Fatalf("%s: golden has %d records, suite produces %d (run -update?)",
					suite.file, len(want), len(live))
			}

			for i := range live {
				if !reflect.DeepEqual(want[i], live[i]) {
					t.Errorf("%s line %d: golden mismatch for input %q (mode %s)\n golden: %+v\n   live: %+v",
						suite.file, i+1, live[i].Input, live[i].Mode, want[i], live[i])
				}
				verifyRecordInvariants(t, suite.file, i+1, want[i])
			}
		})
	}
}

// boundaryMaskLo is the boundary phoneme (ID 0) low word per ADR-0008. It must
// never appear in a feature_stream entry.
const boundaryMaskLo uint64 = 4194304

// verifyRecordInvariants asserts the schema-level invariants documented in
// docs/golden_jsonl_schema.md hold for a stored golden record (independent of
// whether it deep-equals the live result).
func verifyRecordInvariants(t *testing.T, file string, line int, rec golden.Record) {
	t.Helper()
	in := rec.Input

	// (1) feature_stream length == byte length of input.
	if len(rec.FeatureStream) != len(in) {
		t.Errorf("%s line %d: feature_stream len = %d, want %d (byte len of input)",
			file, line, len(rec.FeatureStream), len(in))
	}

	// (5) boundary mask never appears in the stream.
	for i, m := range rec.FeatureStream {
		if m.Lo == boundaryMaskLo && m.Hi == 0 {
			t.Errorf("%s line %d: feature_stream[%d] is the boundary mask; must never be emitted",
				file, line, i)
		}
	}

	causal := rec.Mode == "causal"
	// (4) causal mode ⇒ all-zero feature_stream.
	if causal {
		for i, m := range rec.FeatureStream {
			if m.Lo != 0 || m.Hi != 0 {
				t.Errorf("%s line %d: causal feature_stream[%d] = {lo:%d,hi:%d}, want zero",
					file, line, i, m.Lo, m.Hi)
			}
		}
	}

	for ti, tok := range rec.Tokens {
		// (2) byte-offset span bounds.
		if tok.Span.Start < 0 || tok.Span.End > len(in) || tok.Span.Start > tok.Span.End {
			t.Errorf("%s line %d token %d (%q): span %+v out of bounds [0,%d]",
				file, line, ti, tok.Token, tok.Span, len(in))
		}
		// Surface must equal the byte slice it spans.
		if tok.Span.Start >= 0 && tok.Span.End <= len(in) && tok.Span.Start <= tok.Span.End {
			if got := in[tok.Span.Start:tok.Span.End]; got != tok.Token {
				t.Errorf("%s line %d token %d: surface %q != input[%d:%d]=%q",
					file, line, ti, tok.Token, tok.Span.Start, tok.Span.End, got)
			}
		}

		// (3)/(4) unknown-source tokens carry nothing.
		if tok.Source == "unknown" {
			if len(tok.Phonemes) != 0 || len(tok.Alignment) != 0 {
				t.Errorf("%s line %d token %d (%q): source=unknown but %d phonemes / %d alignment, want 0/0",
					file, line, ti, tok.Token, len(tok.Phonemes), len(tok.Alignment))
			}
		}
		if causal && tok.Source != "unknown" {
			t.Errorf("%s line %d token %d (%q): causal source = %q, want unknown",
				file, line, ti, tok.Token, tok.Source)
		}

		// len(alignment) == len(phonemes) == len(features).
		if len(tok.Alignment) != len(tok.Phonemes) {
			t.Errorf("%s line %d token %d: %d alignment vs %d phonemes",
				file, line, ti, len(tok.Alignment), len(tok.Phonemes))
		}
		if len(tok.Features) != len(tok.Phonemes) {
			t.Errorf("%s line %d token %d: %d features vs %d phonemes",
				file, line, ti, len(tok.Features), len(tok.Phonemes))
		}

		// (6) dict-hit alignment is a contiguous partition of the token span.
		if tok.Source == "dict" && len(tok.Alignment) > 0 {
			if tok.Alignment[0].Start != tok.Span.Start {
				t.Errorf("%s line %d token %d (%q): alignment starts %d, want %d",
					file, line, ti, tok.Token, tok.Alignment[0].Start, tok.Span.Start)
			}
			last := tok.Alignment[len(tok.Alignment)-1]
			if last.End != tok.Span.End {
				t.Errorf("%s line %d token %d (%q): alignment ends %d, want %d",
					file, line, ti, tok.Token, last.End, tok.Span.End)
			}
			for k := 1; k < len(tok.Alignment); k++ {
				if tok.Alignment[k-1].End != tok.Alignment[k].Start {
					t.Errorf("%s line %d token %d (%q): alignment span %d not contiguous",
						file, line, ti, tok.Token, k)
				}
			}
		}
	}
}

// TestGoldenRoundTrip confirms a Record survives marshal→read unchanged, so the
// JSONL representation is lossless for every field the schema captures.
func TestGoldenRoundTrip(t *testing.T) {
	for _, suite := range goldenSuites() {
		for _, c := range suite.cases {
			rec := processCase(t, c)
			var buf bytes.Buffer
			if err := golden.WriteAll(&buf, []golden.Record{rec}); err != nil {
				t.Fatalf("WriteAll: %v", err)
			}
			got, err := golden.ReadAll(&buf)
			if err != nil {
				t.Fatalf("ReadAll: %v", err)
			}
			if len(got) != 1 {
				t.Fatalf("round trip produced %d records, want 1", len(got))
			}
			if !reflect.DeepEqual(got[0], rec) {
				t.Errorf("round trip mismatch for %q (mode %v)\n in: %+v\nout: %+v",
					c.input, c.mode, rec, got[0])
			}
		}
	}
}
