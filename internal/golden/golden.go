// Package golden converts a g2p.Result into a stable, JSON-serializable golden
// record and back, and reads/writes newline-delimited JSON (JSONL) golden
// fixtures used by the conformance and regression tests.
//
// This is an internal package and carries no API stability guarantee. The
// golden record schema it (de)serializes is documented in
// docs/golden_jsonl_schema.md and versioned by Result.SchemaVersion
// ("gofonix-result-v0.1"). The record deliberately exposes only neutral
// phoneme IDs and {lo,hi} feature words — never RAW ARPAbet symbols
// (ADR-0002, ADR-0008).
package golden

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/NK8007/gofonix/pkg/g2p"
	"github.com/NK8007/gofonix/pkg/phoneme"
)

// Mode strings used in golden records. These are the canonical, stable
// spellings of g2p.Mode for serialization (ADR-0003); they are decoupled from
// the Go enum's integer values so a future reordering cannot silently rewrite
// every golden file.
const (
	modeBatch  = "batch"
	modeOracle = "oracle"
	modeCausal = "causal"
)

// Mask is the JSON form of a phoneme.FeatureMask: the two raw 64-bit words.
// Hi is reserved (always 0 in the v0.1-en schema) but is serialized explicitly
// so the record is self-describing and future-proof (ADR-0008).
type Mask struct {
	Lo uint64 `json:"lo"`
	Hi uint64 `json:"hi"`
}

// Span is the JSON form of a g2p.ByteSpan: a half-open [start,end) byte range.
type Span struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// Phoneme is the JSON form of a phoneme.Phoneme. Only the neutral integer ID is
// serialized; the debug-only IPA rendering is intentionally omitted so it can
// never become a serialization contract (ADR-0008, phoneme.Phoneme docs).
type Phoneme struct {
	ID int `json:"id"`
}

// Token is the JSON form of a g2p.TokenResult.
type Token struct {
	Token     string    `json:"token"`
	Span      Span      `json:"span"`
	Kind      string    `json:"kind"`
	Source    string    `json:"source"`
	Variant   int       `json:"variant"`
	Phonemes  []Phoneme `json:"phonemes"`
	Alignment []Span    `json:"alignment"`
	Features  []Mask    `json:"features"`
}

// Record is the full JSON golden record for one Process call. Field order here
// is also the key order json.Marshal emits, giving stable, diffable fixtures.
type Record struct {
	SchemaVersion        string  `json:"schema_version"`
	GofonixVersion       string  `json:"gofonix_version"`
	DictionaryID         string  `json:"dictionary_id"`
	DictionaryChecksum   string  `json:"dictionary_checksum"`
	FeatureSchemaVersion string  `json:"feature_schema_version"`
	NormalizerVersion    string  `json:"normalizer_version"`
	TokenizerVersion     string  `json:"tokenizer_version"`
	OOVPolicy            string  `json:"oov_policy"`
	Language             string  `json:"language"`
	Mode                 string  `json:"mode"`
	Input                string  `json:"input"`
	FeatureStream        []Mask  `json:"feature_stream"`
	Tokens               []Token `json:"tokens"`
}

// kindString maps a g2p.TokenKind to its stable golden spelling.
func kindString(k g2p.TokenKind) string {
	switch k {
	case g2p.KindWord:
		return "word"
	case g2p.KindWhitespace:
		return "whitespace"
	case g2p.KindPunctuation:
		return "punctuation"
	case g2p.KindNumber:
		return "number"
	case g2p.KindSymbol:
		return "symbol"
	case g2p.KindUnknown:
		return "unknown"
	default:
		return fmt.Sprintf("kind(%d)", int(k))
	}
}

// sourceString maps a g2p.Source to its stable golden spelling.
func sourceString(s g2p.Source) string {
	switch s {
	case g2p.SourceDict:
		return "dict"
	case g2p.SourceRuleFallback:
		return "rule_fallback"
	case g2p.SourceUnknown:
		return "unknown"
	default:
		return fmt.Sprintf("source(%d)", int(s))
	}
}

// ModeString maps a g2p.Mode to its stable golden spelling. It is exported so
// test harnesses can build a Record's Mode field consistently.
func ModeString(m g2p.Mode) string {
	switch m {
	case g2p.ModeBatch:
		return modeBatch
	case g2p.ModeOracle:
		return modeOracle
	case g2p.ModeCausal:
		return modeCausal
	default:
		return fmt.Sprintf("mode(%d)", int(m))
	}
}

// ParseMode maps a stable golden Mode spelling back to a g2p.Mode. The second
// result is false for an unrecognized string.
func ParseMode(s string) (g2p.Mode, bool) {
	switch s {
	case modeBatch:
		return g2p.ModeBatch, true
	case modeOracle:
		return g2p.ModeOracle, true
	case modeCausal:
		return g2p.ModeCausal, true
	default:
		return 0, false
	}
}

// maskOf converts a phoneme.FeatureMask to its JSON form.
func maskOf(m phoneme.FeatureMask) Mask { return Mask{Lo: m.Lo(), Hi: m.Hi()} }

// FromResult converts a g2p.Result into its golden Record. The conversion is
// total and lossless for everything the schema captures; it never inspects RAW
// ARPAbet, only neutral phoneme IDs and {lo,hi} feature words.
func FromResult(r g2p.Result) Record {
	stream := make([]Mask, len(r.FeatureStream))
	for i, m := range r.FeatureStream {
		stream[i] = maskOf(m)
	}

	tokens := make([]Token, len(r.Tokens))
	for i, tr := range r.Tokens {
		pr := tr.Pronunciation

		phs := make([]Phoneme, len(pr.Phonemes))
		for j, p := range pr.Phonemes {
			phs[j] = Phoneme{ID: p.ID}
		}

		aln := make([]Span, len(pr.Alignment))
		for j, s := range pr.Alignment {
			aln[j] = Span{Start: s.Start, End: s.End}
		}

		// Per-phoneme feature masks let a golden record be checked without
		// re-deriving the v0.1-en table. There is one entry per phoneme; the
		// boundary phoneme (ID 0) is never produced on the v0.1 dict path so
		// no record exercises it, but FeatureFor(0) would still resolve.
		feats := make([]Mask, len(pr.Phonemes))
		for j, p := range pr.Phonemes {
			feats[j] = maskOf(g2p.FeatureForPhoneme(p))
		}

		tokens[i] = Token{
			Token:     tr.Token,
			Span:      Span{Start: tr.Span.Start, End: tr.Span.End},
			Kind:      kindString(tr.Kind),
			Source:    sourceString(pr.Source),
			Variant:   pr.Variant,
			Phonemes:  phs,
			Alignment: aln,
			Features:  feats,
		}
	}

	return Record{
		SchemaVersion:        r.SchemaVersion,
		GofonixVersion:       r.Metadata.GofonixVersion,
		DictionaryID:         r.Metadata.DictionaryID,
		DictionaryChecksum:   r.Metadata.DictionaryChecksum,
		FeatureSchemaVersion: r.Metadata.FeatureSchemaVersion,
		NormalizerVersion:    r.Metadata.NormalizerVersion,
		TokenizerVersion:     r.Metadata.TokenizerVersion,
		OOVPolicy:            r.Metadata.OOVPolicy,
		Language:             r.Language,
		Mode:                 ModeString(r.Mode),
		Input:                r.Input,
		FeatureStream:        stream,
		Tokens:               tokens,
	}
}

// MarshalLine serializes a Record as a single compact JSON line (no trailing
// newline). Keys follow the struct field order, so output is stable and
// diff-friendly across runs.
func MarshalLine(rec Record) ([]byte, error) {
	return json.Marshal(rec)
}

// ReadAll reads a JSONL stream and returns one Record per non-empty line.
// Blank lines are skipped. A malformed line yields an error naming its 1-based
// line number.
func ReadAll(r io.Reader) ([]Record, error) {
	var out []Record
	sc := bufio.NewScanner(r)
	// Allow long lines: a record with a large input/feature stream can exceed
	// the default 64 KiB token limit.
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	line := 0
	for sc.Scan() {
		line++
		raw := strings.TrimSpace(sc.Text())
		if raw == "" {
			continue
		}
		var rec Record
		if err := json.Unmarshal([]byte(raw), &rec); err != nil {
			return nil, fmt.Errorf("golden: line %d: %w", line, err)
		}
		out = append(out, rec)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// WriteAll writes records as a JSONL stream, one compact JSON object per line,
// each terminated by a single '\n'. Output is deterministic for a given input.
func WriteAll(w io.Writer, recs []Record) error {
	bw := bufio.NewWriter(w)
	for _, rec := range recs {
		b, err := MarshalLine(rec)
		if err != nil {
			return err
		}
		if _, err := bw.Write(b); err != nil {
			return err
		}
		if err := bw.WriteByte('\n'); err != nil {
			return err
		}
	}
	return bw.Flush()
}
