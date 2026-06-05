package phoneme

// Phoneme is a language-neutral phonological unit. The neutral integer ID is
// the public identity (ADR-0008): IDs 1-39 are the v0.1 English inventory, ID 0
// is the universal boundary / non-speech symbol, and IDs 40+ are reserved for
// future expansion and other languages.
//
// IPA is a human-readable debug rendering only. It is not a stable contract and
// must not be used as a serialization key or for cross-version comparison.
type Phoneme struct {
	// ID is the neutral phoneme identity. See ADR-0008 for the v0.1-en
	// inventory and the ID namespace rules.
	ID int

	// IPA is a debug-only rendering of the phoneme (e.g. "k", "æ"). It carries
	// no stability guarantee.
	IPA string
}
