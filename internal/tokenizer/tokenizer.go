package tokenizer

import (
	"unicode"
	"unicode/utf8"
)

// Version is the tokenizer version tag recorded in result metadata (ADR-0006).
const Version = "v0.1"

// asciiDigit reports whether r is an ASCII digit 0-9.
func asciiDigit(r rune) bool { return r >= '0' && r <= '9' }

// isApostrophe reports whether r is an ASCII apostrophe (U+0027) or a right
// single quotation mark (U+2019). Both are contraction apostrophes per ADR-0006.
func isApostrophe(r rune) bool { return r == '\'' || r == '\u2019' }

// Tokenize scans input left-to-right and returns the deterministic token
// sequence defined by ADR-0006. It never returns an error: invalid UTF-8 bytes
// are emitted as single-byte KindUnknown tokens.
func Tokenize(input string) []Token {
	tokens := make([]Token, 0, len(input)/2+1)
	i := 0
	n := len(input)

	for i < n {
		r, size := utf8.DecodeRuneInString(input[i:])

		// Invalid UTF-8: emit one single-byte unknown token and advance one byte.
		if r == utf8.RuneError && size == 1 {
			tokens = append(tokens, Token{
				Surface: input[i : i+1],
				Start:   i,
				End:     i + 1,
				Kind:    KindUnknown,
			})
			i++
			continue
		}

		switch {
		case startsAlphanumericRun(input, i, r, size):
			tok := scanAlphanumericRun(input, i)
			tokens = append(tokens, tok)
			i = tok.End

		case unicode.IsSpace(r):
			tok := scanWhitespace(input, i)
			tokens = append(tokens, tok)
			i = tok.End

		case isHyphenMinus(r):
			tok := scanHyphenRun(input, i)
			tokens = append(tokens, tok)
			i = tok.End

		case unicode.IsPunct(r):
			tokens = append(tokens, Token{
				Surface: input[i : i+size],
				Start:   i,
				End:     i + size,
				Kind:    KindPunctuation,
			})
			i += size

		case unicode.IsSymbol(r):
			tokens = append(tokens, Token{
				Surface: input[i : i+size],
				Start:   i,
				End:     i + size,
				Kind:    KindSymbol,
			})
			i += size

		default:
			// Anything else (e.g. control characters) is unknown, one rune.
			tokens = append(tokens, Token{
				Surface: input[i : i+size],
				Start:   i,
				End:     i + size,
				Kind:    KindUnknown,
			})
			i += size
		}
	}

	return tokens
}

// isHyphenMinus reports whether r is the ASCII hyphen-minus U+002D. Only this
// code point splits words and groups into a hyphen-run punctuation token
// (ADR-0006). Other dashes are ordinary punctuation.
func isHyphenMinus(r rune) bool { return r == '-' }

// startsAlphanumericRun reports whether position i begins an alphanumeric run
// under the unified scanner rule (ADR-0006 case 1): an ASCII letter, an ASCII
// digit, or a sign (+/-) immediately followed by a digit.
func startsAlphanumericRun(input string, i int, r rune, size int) bool {
	if unicode.IsLetter(r) || asciiDigit(r) {
		return true
	}
	if r == '+' || r == '-' {
		// Sign begins a run only if immediately followed by a digit.
		j := i + size
		if j < len(input) {
			nr, _ := utf8.DecodeRuneInString(input[j:])
			return asciiDigit(nr)
		}
	}
	return false
}

// scanAlphanumericRun consumes the maximal contiguous alphanumeric run starting
// at i and classifies it once, per ADR-0006 case 1.
//
//   - If the run contains at least one ASCII letter, the entire run is a single
//     KindWord token (word-internal contraction apostrophes are absorbed).
//   - Otherwise (digits only, with optional leading sign and at most one
//     internal '.' or ',') it is a KindNumber token.
//
// The classification first determines whether the run is letter-bearing (a
// word) or digits-only (a number), because the two cases consume slightly
// different characters (apostrophes vs. a numeric separator).
func scanAlphanumericRun(input string, i int) Token {
	// Decide word vs. number by looking at the leading character. A sign or a
	// digit with no following letter inside the contiguous digit/letter run
	// means number; any letter present means word. We probe the run by scanning
	// letters/digits first.
	start := i
	n := len(input)

	// Determine whether this run (ignoring sign) contains a letter, by scanning
	// the contiguous ASCII letter/digit/apostrophe span. Apostrophes only count
	// as part of a word, so we evaluate letter presence over letters+digits.
	signLen := 0
	if input[i] == '+' || input[i] == '-' {
		signLen = 1
	}

	hasLetter := false
	j := i + signLen
	for j < n {
		r, size := utf8.DecodeRuneInString(input[j:])
		if r == utf8.RuneError && size == 1 {
			break
		}
		if unicode.IsLetter(r) {
			hasLetter = true
			j += size
			continue
		}
		if asciiDigit(r) {
			j += size
			continue
		}
		break
	}

	if hasLetter {
		return scanWord(input, start)
	}
	return scanNumber(input, start)
}

// scanWord consumes a maximal KindWord run starting at i. A word is a run of
// ASCII letters and digits, plus word-internal apostrophes (ASCII U+0027 or
// U+2019) that lie strictly between two Unicode letters (ADR-0006).
//
// A leading sign is never part of a word (signs are only absorbed by numbers),
// so scanWord is only entered for runs whose first significant character is a
// letter or digit.
func scanWord(input string, i int) Token {
	n := len(input)
	pos := i

	for pos < n {
		r, size := utf8.DecodeRuneInString(input[pos:])

		if unicode.IsLetter(r) || asciiDigit(r) {
			pos += size
			continue
		}

		if isApostrophe(r) {
			// Word-internal only if flanked by letters on BOTH sides.
			// Left side: previous rune must be a letter.
			if !prevRuneIsLetter(input, pos) {
				break
			}
			// Right side: next rune must be a letter.
			next := pos + size
			if next >= n {
				break
			}
			nr, _ := utf8.DecodeRuneInString(input[next:])
			if !unicode.IsLetter(nr) {
				break
			}
			pos += size
			continue
		}

		break
	}

	return Token{
		Surface: input[i:pos],
		Start:   i,
		End:     pos,
		Kind:    KindWord,
	}
}

// prevRuneIsLetter reports whether the rune ending at byte offset pos is a
// Unicode letter.
func prevRuneIsLetter(input string, pos int) bool {
	if pos == 0 {
		return false
	}
	r, _ := utf8.DecodeLastRuneInString(input[:pos])
	return unicode.IsLetter(r)
}

// scanNumber consumes a KindNumber token starting at i per the ADR-0006 number
// grammar: sign? digits ( sep digits )?, with at most one internal '.' or ','.
//
// scanNumber is only entered for digits-only runs (no letters). The sign is
// part of the number only because the caller verified it is immediately
// followed by a digit.
func scanNumber(input string, i int) Token {
	n := len(input)
	pos := i

	if pos < n && (input[pos] == '+' || input[pos] == '-') {
		pos++
	}

	// First digit group (guaranteed at least one digit by the caller).
	for pos < n && input[pos] >= '0' && input[pos] <= '9' {
		pos++
	}

	// Optional single internal separator followed by more digits.
	if pos < n && (input[pos] == '.' || input[pos] == ',') {
		// The separator is only part of the number if a digit follows it.
		if pos+1 < n && input[pos+1] >= '0' && input[pos+1] <= '9' {
			pos++ // consume separator
			for pos < n && input[pos] >= '0' && input[pos] <= '9' {
				pos++
			}
		}
	}

	return Token{
		Surface: input[i:pos],
		Start:   i,
		End:     pos,
		Kind:    KindNumber,
	}
}

// scanWhitespace consumes the maximal Unicode White_Space run starting at i.
func scanWhitespace(input string, i int) Token {
	n := len(input)
	pos := i
	for pos < n {
		r, size := utf8.DecodeRuneInString(input[pos:])
		if r == utf8.RuneError && size == 1 {
			break
		}
		if !unicode.IsSpace(r) {
			break
		}
		pos += size
	}
	return Token{
		Surface: input[i:pos],
		Start:   i,
		End:     pos,
		Kind:    KindWhitespace,
	}
}

// scanHyphenRun consumes a maximal run of ASCII hyphen-minus U+002D as a single
// KindPunctuation token (ADR-0006 grouped-hyphen rule).
func scanHyphenRun(input string, i int) Token {
	n := len(input)
	pos := i
	for pos < n && input[pos] == '-' {
		pos++
	}
	return Token{
		Surface: input[i:pos],
		Start:   i,
		End:     pos,
		Kind:    KindPunctuation,
	}
}
