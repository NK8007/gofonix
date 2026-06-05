package dict

import (
	"strconv"
	"strings"
)

// parseLine parses a single CMUdict line per the ADR-0004 parsing spec.
//
// Return values:
//
//   - key: the lowercased WORD field (ADR-0006 lookup-key policy).
//   - phones: the raw, space-separated ARPAbet symbols (no mapping, no stress
//     stripping; e.g. ["K", "AE1", "T"]).
//   - variant: 0 for an unnumbered (base) entry; N >= 1 for a "WORD(N)"
//     alternate entry.
//   - ok: false when the line must be skipped deterministically (comment, blank,
//     or malformed). The same input always produces the same skip decision.
//
// Rules:
//   - Lines beginning with ";;;" are comments and are skipped.
//   - Blank / whitespace-only lines are skipped.
//   - The format is "WORD  PH1 PH2 ..." with one or more whitespace characters
//     separating the word from the phonemes and separating the phonemes.
//   - A variant entry has the form "WORD(N)  PH1 ..." where N is an integer >= 1.
//   - A line with no phonemes, or a malformed variant suffix, is skipped.
func parseLine(line string) (key string, phones []string, variant int, ok bool) {
	// Comment lines begin with ";;;".
	if strings.HasPrefix(line, ";;;") {
		return "", nil, 0, false
	}

	// Blank or whitespace-only lines are skipped.
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return "", nil, 0, false
	}

	// Split into fields on any run of whitespace. The first field is the WORD
	// (possibly with a "(N)" variant suffix); the remaining fields are phonemes.
	fields := strings.Fields(trimmed)
	if len(fields) < 2 {
		// No phonemes present: deterministic skip.
		return "", nil, 0, false
	}

	word := fields[0]
	phones = fields[1:]

	// Detect and strip a "(N)" variant suffix.
	variant = 0
	if i := strings.IndexByte(word, '('); i >= 0 {
		// Must end with ')' and have a non-empty interior between '(' and ')'.
		if !strings.HasSuffix(word, ")") || i+1 >= len(word)-1 {
			return "", nil, 0, false
		}
		inner := word[i+1 : len(word)-1]
		n, err := strconv.Atoi(inner)
		if err != nil || n < 1 {
			// Malformed or non-positive variant number: deterministic skip.
			return "", nil, 0, false
		}
		variant = n
		word = word[:i]
	}

	if word == "" {
		return "", nil, 0, false
	}

	// Normalize the key to lowercase (ADR-0006 lookup-key policy).
	key = strings.ToLower(word)

	return key, phones, variant, true
}
