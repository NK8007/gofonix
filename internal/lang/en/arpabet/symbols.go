// Package arpabet is the internal English ARPAbet bridge (ADR-0008; ARPAbet stays internal).
//
// ARPAbet lives ONLY here: this package maps stress-stripped CMUdict ARPAbet
// symbols to neutral phoneme.Phoneme values (IDs 1-39 per the v0.1-en
// inventory) and exposes the v0.1-en FeatureMask for each neutral ID. No
// ARPAbet string ever escapes into the public g2p API; callers receive neutral
// phonemes only. This package carries no API stability guarantee.
package arpabet

// symbolToID maps a stress-stripped ARPAbet symbol to its neutral phoneme ID.
// IDs are per ADR-0008 v0.1-en inventory (IDs 1-39). ID 0 (boundary) is never
// produced from an ARPAbet symbol.
var symbolToID = map[string]int{
	"AA": 1, "AE": 2, "AH": 3, "AO": 4, "AW": 5, "AY": 6,
	"B": 7, "CH": 8, "D": 9, "DH": 10,
	"EH": 11, "ER": 12, "EY": 13,
	"F": 14, "G": 15, "HH": 16,
	"IH": 22, "IY": 18, "JH": 19, "K": 20,
	"L": 21, "M": 17, "N": 23, "NG": 24,
	"OW": 25, "OY": 26, "P": 27, "R": 28,
	"S": 29, "SH": 30, "T": 31, "TH": 32,
	"UH": 33, "UW": 34, "V": 35, "W": 36,
	"Y": 37, "Z": 38, "ZH": 39,
}
