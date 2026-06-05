package fallback

import "strings"

// Normalize applies the ADR-0009 normalization pipeline to a surface form and
// reports whether the result is eligible for fallback matching.
//
// The pipeline runs, in order:
//  1. lowercase (strict ASCII fold: only [A-Z] is mapped to [a-z]; no Unicode
//     case folding is performed);
//  2. if any decimal digit [0-9] is present -> ("", false);
//  3. if any byte is not an ASCII letter [a-z] (post-fold) nor an ASCII
//     apostrophe U+0027 -> ("", false); any byte >= 0x80 (i.e. any non-ASCII
//     code point) is rejected here;
//  4. strip all ASCII apostrophes U+0027 (e.g. "don't" -> "dont");
//  5. if the result is empty -> ("", false).
//
// On success it returns the normalized, apostrophe-stripped, lowercase letter
// sequence and true. Curly apostrophes (U+2018/U+2019) and any non-ASCII
// characters such as 'é' are rejected at step 3.
//
// Lowercasing is performed via a strict ASCII fold (byte + 32 for [A-Z])
// rather than strings.ToLower. strings.ToLower applies Unicode-aware case
// folding, which can collapse certain non-ASCII code points (e.g. U+0130
// LATIN CAPITAL LETTER I WITH DOT ABOVE) toward ASCII before the byte
// validation runs, leaking non-ASCII surface forms through the ASCII-only
// activation predicate. The byte-wise fold guarantees that only genuine
// ASCII input can ever survive.
func Normalize(input string) (string, bool) {
	// Steps 1-3 in a single byte scan. ASCII-only validity means we can scan
	// bytes directly; any multi-byte rune contains bytes >= 0x80, which are
	// rejected at step 3.
	folded := make([]byte, 0, len(input))
	for i := 0; i < len(input); i++ {
		c := input[i]
		if c >= 0x80 {
			return "", false // step 3: non-ASCII byte
		}
		if c >= '0' && c <= '9' {
			return "", false // step 2: digit present
		}
		if c >= 'A' && c <= 'Z' {
			c += 32 // step 1: strict ASCII fold [A-Z] -> [a-z]
		}
		if !((c >= 'a' && c <= 'z') || c == '\'') {
			return "", false // step 3: disallowed character
		}
		folded = append(folded, c)
	}

	// Step 4: strip all ASCII apostrophes.
	stripped := strings.ReplaceAll(string(folded), "'", "")

	// Step 5: empty after stripping.
	if stripped == "" {
		return "", false
	}

	return stripped, true
}

// Match runs the deterministic greedy longest-match left-to-right scan over an
// already-normalized string (ADR-0009 "Longest-Match Algorithm"). The input is
// assumed to be the output of Normalize (lowercase, letters only).
//
// At each position it selects the longest table grapheme that matches; ties in
// length are broken by choosing the alphabetically-first grapheme key, making
// the result independent of map iteration order. It returns the accumulated
// ARPAbet symbol sequence and true on success. If no grapheme matches at some
// position (only possible for a non-letter character that bypassed Normalize),
// it returns (nil, false).
func Match(normalized string) ([]string, bool) {
	if normalized == "" {
		return nil, false
	}

	out := make([]string, 0, len(normalized))
	for i := 0; i < len(normalized); {
		bestGrapheme := ""
		var bestSymbols []string

		// Probe candidate grapheme lengths from longest to shortest. The first
		// length that yields any match gives the longest match; among multiple
		// graphemes of that same length we keep the alphabetically-first key.
		// (Within a fixed length there is at most one substring of the input,
		// so the alphabetical tie-break is only meaningful across the table's
		// own equal-length keys; probing the single substring is sufficient and
		// the alphabetical rule is preserved because that substring is itself
		// the unique candidate of that length.)
		maxLen := maxGraphemeLen
		if remaining := len(normalized) - i; remaining < maxLen {
			maxLen = remaining
		}
		for l := maxLen; l >= 1; l-- {
			candidate := normalized[i : i+l]
			if syms, ok := ruleIndex[candidate]; ok {
				bestGrapheme = candidate
				bestSymbols = syms
				break
			}
		}

		if bestGrapheme == "" {
			return nil, false // no grapheme matches: decline
		}

		out = append(out, bestSymbols...)
		i += len(bestGrapheme)
	}

	return out, true
}

// Pronounce is the convenience entry point: it normalizes the surface form and,
// if eligible, runs the longest-match scan. It returns (symbols, true) for a
// resolved fallback pronunciation, or (nil, false) when the surface form is
// declined at normalization or matching.
func Pronounce(surface string) ([]string, bool) {
	normalized, ok := Normalize(surface)
	if !ok {
		return nil, false
	}
	return Match(normalized)
}
