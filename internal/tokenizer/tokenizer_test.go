package tokenizer

import (
	"fmt"
	"testing"
)

// expTok is one expected token: kind plus [Start,End) byte span.
type expTok struct {
	kind  Kind
	start int
	end   int
}

func kindName(k Kind) string {
	switch k {
	case KindWord:
		return "Word"
	case KindWhitespace:
		return "Whitespace"
	case KindPunctuation:
		return "Punctuation"
	case KindNumber:
		return "Number"
	case KindSymbol:
		return "Symbol"
	case KindUnknown:
		return "Unknown"
	default:
		return fmt.Sprintf("Kind(%d)", int(k))
	}
}

// TestTokenizeTable pins exact tokens + spans for the ADR-0006 golden inputs.
func TestTokenizeTable(t *testing.T) {
	u2019 := "\u2019" // right single quotation mark (3 bytes)

	cases := []struct {
		name  string
		input string
		want  []expTok
	}{
		{"cat", "cat", []expTok{{KindWord, 0, 3}}},
		{"contraction-ascii", "don't", []expTok{{KindWord, 0, 5}}},
		// "it" + U+2019 + "s": 2 + 3 + 1 = 6 bytes -> one word [0,6).
		{"contraction-u2019", "it" + u2019 + "s", []expTok{{KindWord, 0, 6}}},
		{"quoted-word", "'hello'", []expTok{
			{KindPunctuation, 0, 1},
			{KindWord, 1, 6},
			{KindPunctuation, 6, 7},
		}},
		{"hyphen-split", "well-known", []expTok{
			{KindWord, 0, 4},
			{KindPunctuation, 4, 5},
			{KindWord, 5, 10},
		}},
		{"double-hyphen", "well--known", []expTok{
			{KindWord, 0, 4},
			{KindPunctuation, 4, 6},
			{KindWord, 6, 11},
		}},
		{"leading-hyphen-letter", "-x", []expTok{
			{KindPunctuation, 0, 1},
			{KindWord, 1, 2},
		}},
		{"signed-number", "-5", []expTok{{KindNumber, 0, 2}}},
		{"int", "123", []expTok{{KindNumber, 0, 3}}},
		{"decimal", "3.14", []expTok{{KindNumber, 0, 4}}},
		{"ordinal-3rd", "3rd", []expTok{{KindWord, 0, 3}}},
		{"ordinal-21st", "21st", []expTok{{KindWord, 0, 4}}},
		{"mp3", "MP3", []expTok{{KindWord, 0, 3}}},
		{"h2o", "H2O", []expTok{{KindWord, 0, 3}}},
		{"space", " ", []expTok{{KindWhitespace, 0, 1}}},
		{"period", ".", []expTok{{KindPunctuation, 0, 1}}},
		{"invalid-utf8", "\xff", []expTok{{KindUnknown, 0, 1}}},
		// Unicode letter word runs (ADR-0006: Lu, Ll, Lt, Lm, Lo).
		// "café" = c(1)+a(1)+f(1)+é(2) = 5 bytes.
		{"unicode-word", "café", []expTok{{KindWord, 0, 5}}},
		// "naïve" = n(1)+a(1)+ï(2)+v(1)+e(1) = 6 bytes.
		{"unicode-mixed", "naïve", []expTok{{KindWord, 0, 6}}},
		// "München" = M(1)+ü(2)+n(1)+c(1)+h(1)+e(1)+n(1) = 8 bytes.
		{"unicode-munchen", "München", []expTok{{KindWord, 0, 8}}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Tokenize(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("input %q: got %d tokens, want %d\n got=%s",
					tc.input, len(got), len(tc.want), dump(got))
			}
			for i := range tc.want {
				w := tc.want[i]
				g := got[i]
				if g.Kind != w.kind || g.Start != w.start || g.End != w.end {
					t.Errorf("input %q token %d: got %s[%d,%d) want %s[%d,%d)",
						tc.input, i,
						kindName(g.Kind), g.Start, g.End,
						kindName(w.kind), w.start, w.end)
				}
				// Surface must equal the original input bytes for the span.
				if g.Surface != tc.input[g.Start:g.End] {
					t.Errorf("input %q token %d: surface %q != input[%d:%d]=%q",
						tc.input, i, g.Surface, g.Start, g.End,
						tc.input[g.Start:g.End])
				}
			}
		})
	}
}

func dump(toks []Token) string {
	s := ""
	for _, t := range toks {
		s += fmt.Sprintf("%s[%d,%d)=%q ", kindName(t.Kind), t.Start, t.End, t.Surface)
	}
	return s
}

// TestSpanContiguity asserts the tokens fully and contiguously cover the input
// (a partition), which ADR-0007 relies on for byte-aligned FeatureStream.
func TestSpanContiguity(t *testing.T) {
	inputs := []string{
		"hello world", "don't -5 well--known 3.14 .", "'a'-b", "\xffx\xff",
		"", " \t\n ",
	}
	for _, in := range inputs {
		toks := Tokenize(in)
		prev := 0
		for i, tk := range toks {
			if tk.Start != prev {
				t.Errorf("input %q token %d starts at %d, expected %d (gap/overlap)",
					in, i, tk.Start, prev)
			}
			if tk.Start > tk.End {
				t.Errorf("input %q token %d has Start>End [%d,%d)", in, i, tk.Start, tk.End)
			}
			prev = tk.End
		}
		if prev != len(in) {
			t.Errorf("input %q: tokens cover up to %d, want %d", in, prev, len(in))
		}
	}
}
