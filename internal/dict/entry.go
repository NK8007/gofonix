package dict

// Entry is the stored dictionary record for a single lookup key (ADR-0004).
//
// The phoneme symbols are RAW ARPAbet strings exactly as they appear in the
// CMUdict source, including stress digits (e.g. "AE1", "IH0"). This package
// performs NO mapping to neutral phoneme.Phoneme values and NO stress
// stripping; both are done by the internal/lang/en/arpabet bridge.
type Entry struct {
	// Canonical is the phoneme sequence of the unnumbered (base) CMUdict entry
	// for the key — the variant-0 pronunciation. This is the only pronunciation
	// ever selected by v0.1 lookup (ADR-0004).
	Canonical []string

	// Alternates holds the phoneme sequences of the numbered CMUdict entries
	// (WORD(1), WORD(2), ...), in ascending variant order. They are parsed and
	// stored for future use but are never selected by v0.1 lookup (ADR-0004).
	Alternates [][]string
}
