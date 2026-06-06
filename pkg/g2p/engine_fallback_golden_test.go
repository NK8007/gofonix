package g2p_test

// Golden support for the OOV fallback corpus (ADR-0009).
//
// This file defines a dedicated golden record type and runner for
// testdata/golden/fallback.jsonl. It is intentionally SEPARATE from the
// shared golden.Record / TestGolden machinery for two reasons:
//
//  1. The frozen golden files (words/oov/atypical/nonword/causal) do
//     NOT carry a fallback_rules_version field, and they must not be modified
//     or regenerated here. Adding the field to the shared
//     golden.Record would break their exact-match comparison.
//  2. ADR-0009 requires fallback golden records to pin FallbackRulesVersion
//     byte-for-byte. The fbRecord type below adds exactly that one field,
//     positioned right after oov_policy, while reusing the neutral
//     phoneme-ID / {lo,hi} feature-word representation (no RAW ARPAbet).
//
// The corpus is generated PROGRAMMATICALLY from live Process output: run
//
//	go test ./pkg/g2p/ -run TestGenerateFallbackGolden -update
//
// (or GOFONIX_UPDATE_GOLDEN=1) to (re)write testdata/golden/fallback.jsonl.
// A normal `go test` run never rewrites it and instead compares against it.

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/NK8007/gofonix/internal/golden"
	"github.com/NK8007/gofonix/pkg/g2p"
)

// updateFallbackGolden mirrors the -update / GOFONIX_UPDATE_GOLDEN convention
// used by TestGolden so the fallback corpus is regenerated the same way. The
// flag is registered lazily to avoid clashing with the -update flag already
// declared in golden_test.go (same test binary).
func fallbackUpdateEnabled() bool {
	if f := flag.Lookup("update"); f != nil {
		if bf, ok := f.Value.(interface{ Get() any }); ok {
			if v, ok := bf.Get().(bool); ok && v {
				return true
			}
		}
	}
	return os.Getenv("GOFONIX_UPDATE_GOLDEN") == "1"
}

// fbMask is the JSON form of an opaque phoneme.FeatureMask (two 64-bit words),
// identical to golden.Mask.
type fbMask struct {
	Lo uint64 `json:"lo"`
	Hi uint64 `json:"hi"`
}

// fbSpan is a half-open [start,end) byte range, identical to golden.Span.
type fbSpan struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// fbPhoneme carries only the neutral phoneme ID (never RAW ARPAbet).
type fbPhoneme struct {
	ID int `json:"id"`
}

// fbToken is the JSON form of a g2p.TokenResult.
type fbToken struct {
	Token     string      `json:"token"`
	Span      fbSpan      `json:"span"`
	Kind      string      `json:"kind"`
	Source    string      `json:"source"`
	Variant   int         `json:"variant"`
	Phonemes  []fbPhoneme `json:"phonemes"`
	Alignment []fbSpan    `json:"alignment"`
	Features  []fbMask    `json:"features"`
}

// fbRecord is the golden record for the fallback corpus. Field order is the
// json.Marshal key order, giving stable diffable fixtures. It matches the
// frozen golden schema and adds fallback_rules_version immediately after
// oov_policy (ADR-0009 metadata requirement).
type fbRecord struct {
	SchemaVersion        string    `json:"schema_version"`
	GofonixVersion       string    `json:"gofonix_version"`
	DictionaryID         string    `json:"dictionary_id"`
	DictionaryChecksum   string    `json:"dictionary_checksum"`
	FeatureSchemaVersion string    `json:"feature_schema_version"`
	NormalizerVersion    string    `json:"normalizer_version"`
	TokenizerVersion     string    `json:"tokenizer_version"`
	OOVPolicy            string    `json:"oov_policy"`
	FallbackRulesVersion string    `json:"fallback_rules_version"`
	Language             string    `json:"language"`
	Mode                 string    `json:"mode"`
	Input                string    `json:"input"`
	FeatureStream        []fbMask  `json:"feature_stream"`
	Tokens               []fbToken `json:"tokens"`
}

// fbCase is one (input, mode) pair for the fallback corpus, in file line order.
type fbCase struct {
	input string
	mode  g2p.Mode
}

// fallbackGoldenCases is the authoritative ordered list of fallback golden
// cases (ADR-0009 Test Strategy). Order == line order in fallback.jsonl.
func fallbackGoldenCases() []fbCase {
	const rsquo = "\u2019" // U+2019 right single quotation mark
	return []fbCase{
		// SourceRuleFallback (letter-only OOV words).
		{"ship", g2p.ModeBatch},
		{"thing", g2p.ModeBatch},
		{"phone", g2p.ModeBatch},
		{"xenon", g2p.ModeBatch},
		{"qwerty", g2p.ModeBatch},
		// SourceUnknown (fallback declined: digits / non-ASCII / curly quote).
		{"MP3", g2p.ModeBatch},
		{"H2O", g2p.ModeBatch},
		{"3rd", g2p.ModeBatch},
		{"21st", g2p.ModeBatch},
		{"caf\u00e9", g2p.ModeBatch},
		{"it" + rsquo + "s", g2p.ModeBatch},
		// SourceDict regression.
		{"cat", g2p.ModeBatch},
		// Composite: "ship" (rule_fallback) + "." (unknown).
		{"ship.", g2p.ModeBatch},
		// Causal: all-zero stream, source unknown, empty phonemes.
		{"qwerty", g2p.ModeCausal},
	}
}

// fbFromResult converts a live g2p.Result into an fbRecord by reusing the
// canonical golden.FromResult conversion (so the phoneme-ID / feature-word
// derivation is identical to the rest of the corpus) and then attaching the
// fallback_rules_version surfaced in the result metadata.
func fbFromResult(res g2p.Result) fbRecord {
	r := golden.FromResult(res)

	stream := make([]fbMask, len(r.FeatureStream))
	for i, m := range r.FeatureStream {
		stream[i] = fbMask{Lo: m.Lo, Hi: m.Hi}
	}
	toks := make([]fbToken, len(r.Tokens))
	for i, t := range r.Tokens {
		phs := make([]fbPhoneme, len(t.Phonemes))
		for j, p := range t.Phonemes {
			phs[j] = fbPhoneme{ID: p.ID}
		}
		aln := make([]fbSpan, len(t.Alignment))
		for j, s := range t.Alignment {
			aln[j] = fbSpan{Start: s.Start, End: s.End}
		}
		feats := make([]fbMask, len(t.Features))
		for j, m := range t.Features {
			feats[j] = fbMask{Lo: m.Lo, Hi: m.Hi}
		}
		toks[i] = fbToken{
			Token:     t.Token,
			Span:      fbSpan{Start: t.Span.Start, End: t.Span.End},
			Kind:      t.Kind,
			Source:    t.Source,
			Variant:   t.Variant,
			Phonemes:  phs,
			Alignment: aln,
			Features:  feats,
		}
	}
	return fbRecord{
		SchemaVersion:        r.SchemaVersion,
		GofonixVersion:       r.GofonixVersion,
		DictionaryID:         r.DictionaryID,
		DictionaryChecksum:   r.DictionaryChecksum,
		FeatureSchemaVersion: r.FeatureSchemaVersion,
		NormalizerVersion:    r.NormalizerVersion,
		TokenizerVersion:     r.TokenizerVersion,
		OOVPolicy:            r.OOVPolicy,
		FallbackRulesVersion: res.Metadata.FallbackRulesVersion,
		Language:             r.Language,
		Mode:                 r.Mode,
		Input:                r.Input,
		FeatureStream:        stream,
		Tokens:               toks,
	}
}

// fbProcess runs one fallback case through a fresh Engine and returns the live
// fbRecord.
func fbProcess(t *testing.T, c fbCase) fbRecord {
	t.Helper()
	e, err := g2p.New(g2p.Options{Language: "en", Mode: c.mode})
	if err != nil {
		t.Fatalf("New(mode=%v): %v", c.mode, err)
	}
	res, err := e.Process(c.input)
	if err != nil {
		t.Fatalf("Process(%q): %v", c.input, err)
	}
	return fbFromResult(res)
}

// fallbackGoldenPath is the path to the fallback corpus file (pkg/g2p is three
// levels below the module root).
func fallbackGoldenPath() string {
	return filepath.Join("..", "..", "testdata", "golden", "fallback.jsonl")
}

// fbReadAll deserializes every non-empty line of an fbRecord JSONL stream.
func fbReadAll(t *testing.T, raw []byte) []fbRecord {
	t.Helper()
	var out []fbRecord
	for i, line := range bytes.Split(raw, []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var rec fbRecord
		if err := json.Unmarshal(line, &rec); err != nil {
			t.Fatalf("fallback.jsonl line %d: %v", i+1, err)
		}
		out = append(out, rec)
	}
	return out
}

// fbWriteAll serializes records as compact JSONL, one object per line.
func fbWriteAll(t *testing.T, recs []fbRecord) []byte {
	t.Helper()
	var buf bytes.Buffer
	for _, rec := range recs {
		b, err := json.Marshal(rec)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		buf.Write(b)
		buf.WriteByte('\n')
	}
	return buf.Bytes()
}

// TestGenerateFallbackGolden (re)generates testdata/golden/fallback.jsonl from
// live Process output when -update / GOFONIX_UPDATE_GOLDEN=1 is set. Without
// the flag it is a no-op so a normal `go test` run never rewrites the file.
func TestGenerateFallbackGolden(t *testing.T) {
	if !fallbackUpdateEnabled() {
		t.Skip("set -update or GOFONIX_UPDATE_GOLDEN=1 to regenerate fallback.jsonl")
	}
	cases := fallbackGoldenCases()
	recs := make([]fbRecord, len(cases))
	for i, c := range cases {
		recs[i] = fbProcess(t, c)
	}
	path := fallbackGoldenPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, fbWriteAll(t, recs), 0o644); err != nil {
		t.Fatalf("WriteFile %s: %v", path, err)
	}
	t.Logf("wrote %s (%d records)", path, len(recs))
}

// TestGoldenFallbackExactMatch reads fallback.jsonl and, for every record,
// re-runs Process and asserts an exact match on the contract-critical fields:
// per-token source, the neutral phoneme-ID sequence, and len(feature_stream)
// (which must equal len(input) in bytes). It also re-serializes the live
// record and requires byte-for-byte deep equality with the stored record, so
// the corpus is pinned end-to-end (ADR-0009 golden cases).
func TestGoldenFallbackExactMatch(t *testing.T) {
	raw, err := os.ReadFile(fallbackGoldenPath())
	if err != nil {
		t.Fatalf("read fallback.jsonl: %v (run with -update to generate)", err)
	}
	want := fbReadAll(t, raw)
	cases := fallbackGoldenCases()
	if len(want) != len(cases) {
		t.Fatalf("fallback.jsonl has %d records, suite produces %d (run -update?)",
			len(want), len(cases))
	}

	for i, c := range cases {
		live := fbProcess(t, c)
		w := want[i]

		// Whole-record exact match.
		if !reflect.DeepEqual(w, live) {
			t.Errorf("fallback.jsonl line %d (input %q, mode %s): golden mismatch\n golden: %+v\n   live: %+v",
				i+1, c.input, live.Mode, w, live)
			continue
		}

		// Explicit field-level assertions required by ADR-0009.
		if len(w.FeatureStream) != len(c.input) {
			t.Errorf("line %d (%q): feature_stream len = %d, want %d (byte len)",
				i+1, c.input, len(w.FeatureStream), len(c.input))
		}
		if len(w.Tokens) != len(live.Tokens) {
			t.Fatalf("line %d (%q): %d tokens, live %d", i+1, c.input, len(w.Tokens), len(live.Tokens))
		}
		for ti := range w.Tokens {
			wt, lt := w.Tokens[ti], live.Tokens[ti]
			if wt.Source != lt.Source {
				t.Errorf("line %d (%q) token %d: source = %q, live %q",
					i+1, c.input, ti, wt.Source, lt.Source)
			}
			wIDs := phonemeIDs(wt.Phonemes)
			lIDs := phonemeIDs(lt.Phonemes)
			if !reflect.DeepEqual(wIDs, lIDs) {
				t.Errorf("line %d (%q) token %d: phoneme_ids = %v, live %v",
					i+1, c.input, ti, wIDs, lIDs)
			}
		}

		// Mode-specific invariants.
		if w.Mode == "causal" {
			for k, m := range w.FeatureStream {
				if m.Lo != 0 || m.Hi != 0 {
					t.Errorf("line %d (%q): causal feature_stream[%d] = {lo:%d,hi:%d}, want zero",
						i+1, c.input, k, m.Lo, m.Hi)
				}
			}
			for ti, tok := range w.Tokens {
				if tok.Source != "unknown" || len(tok.Phonemes) != 0 {
					t.Errorf("line %d (%q) token %d: causal must be source=unknown, empty phonemes; got %q / %d",
						i+1, c.input, ti, tok.Source, len(tok.Phonemes))
				}
			}
		}
	}
}

// phonemeIDs extracts the neutral ID sequence from a token's phonemes.
func phonemeIDs(ps []fbPhoneme) []int {
	out := make([]int, len(ps))
	for i, p := range ps {
		out[i] = p.ID
	}
	return out
}

// TestDictVsFallbackVsUnknown exercises the three provenance outcomes through
// the live engine in a single test (ADR-0009 provenance tests):
//   - "cat" is a dictionary hit  -> SourceDict
//   - "ship" is a letter-only OOV -> SourceRuleFallback
//   - "MP3" is mixed alphanumeric -> SourceUnknown
func TestDictVsFallbackVsUnknown(t *testing.T) {
	cases := []struct {
		input string
		want  g2p.Source
	}{
		{"cat", g2p.SourceDict},
		{"ship", g2p.SourceRuleFallback},
		{"MP3", g2p.SourceUnknown},
	}
	e, err := g2p.New(g2p.Options{Language: "en", Mode: g2p.ModeBatch})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		res, err := e.Process(c.input)
		if err != nil {
			t.Fatalf("Process(%q): %v", c.input, err)
		}
		// Find the word token (first token whose surface contains a letter
		// run); for these inputs the first token is the relevant one.
		got := res.Tokens[0].Pronunciation.Source
		if got != c.want {
			t.Errorf("Process(%q): token[0] source = %d, want %d", c.input, got, c.want)
		}
		if c.want == g2p.SourceUnknown {
			if len(res.Tokens[0].Pronunciation.Phonemes) != 0 {
				t.Errorf("Process(%q): SourceUnknown must have empty phonemes", c.input)
			}
		} else {
			if len(res.Tokens[0].Pronunciation.Phonemes) == 0 {
				t.Errorf("Process(%q): source %d must have phonemes", c.input, c.want)
			}
		}
	}
}

// TestModeCausalOOVAllZero confirms "qwerty" in ModeCausal yields an all-zero
// feature_stream of length 6 (== byte length), with the token declined to
// SourceUnknown and empty phonemes (ADR-0003 / ADR-0009: causal never applies
// the fallback).
func TestModeCausalOOVAllZero(t *testing.T) {
	const in = "qwerty"
	e, err := g2p.New(g2p.Options{Language: "en", Mode: g2p.ModeCausal})
	if err != nil {
		t.Fatal(err)
	}
	res, err := e.Process(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.FeatureStream) != 6 {
		t.Errorf("feature_stream len = %d, want 6", len(res.FeatureStream))
	}
	if len(res.FeatureStream) != len(in) {
		t.Errorf("feature_stream len = %d, want len(input) = %d", len(res.FeatureStream), len(in))
	}
	for i, m := range res.FeatureStream {
		if !m.IsZero() {
			t.Errorf("feature_stream[%d] non-zero, want all-zero in causal mode", i)
		}
	}
	tok := res.Tokens[0]
	if tok.Pronunciation.Source != g2p.SourceUnknown {
		t.Errorf("source = %d, want SourceUnknown", tok.Pronunciation.Source)
	}
	if len(tok.Pronunciation.Phonemes) != 0 {
		t.Errorf("phonemes = %d, want 0 (causal)", len(tok.Pronunciation.Phonemes))
	}
}
