// Package tokenizer implements the deterministic, left-to-right, maximal-munch
// scanner specified by ADR-0006.
//
// It is an internal package and carries no API stability guarantee. The public
// g2p package adapts these tokens into g2p.TokenResult values.
package tokenizer

// Kind enumerates the token classes produced by the scanner. The values mirror
// g2p.TokenKind (ADR-0001) but are kept internal so the tokenizer has no
// dependency on the public package.
type Kind int

const (
	// KindWord is an alphanumeric run containing at least one Unicode letter,
	// including word-internal contraction apostrophes.
	KindWord Kind = iota
	// KindWhitespace is a maximal run of Unicode White_Space code points.
	KindWhitespace
	// KindPunctuation is a Unicode punctuation token (one code point, except a
	// grouped run of ASCII hyphen-minus).
	KindPunctuation
	// KindNumber is a digits-only run with an optional leading sign and at most
	// one internal '.' or ',' separator.
	KindNumber
	// KindSymbol is a single Unicode symbol code point (Sc, Sk, Sm, So).
	KindSymbol
	// KindUnknown is a single byte that fits no other class, including invalid
	// UTF-8 bytes.
	KindUnknown
)

// Token is a single lexical unit with a half-open [Start, End) byte span into
// the original input string (ADR-0006). Surface is the exact substring
// input[Start:End]; normalization never alters the span.
type Token struct {
	Surface string
	Start   int
	End     int
	Kind    Kind
}
