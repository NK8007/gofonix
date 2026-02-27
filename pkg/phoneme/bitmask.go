package phoneme

// Bitmask constants for phonetic features
const (
	Voiced uint16 = 1 << iota
	Nasal
	Stop
	Fricative
	// TODO: Add more features
)
