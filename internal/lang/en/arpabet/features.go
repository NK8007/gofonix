package arpabet

import "github.com/NK8007/gofonix/pkg/phoneme"

// phonemeFeatures maps neutral phoneme ID to its FeatureMask per ADR-0008
// v0.1-en. The decimal lo word in each trailing comment is the authoritative
// value from the ADR-0008 feature-assignment table and is pinned by tests.
var phonemeFeatures = map[int]phoneme.FeatureMask{
	0:  phoneme.FromBits(phoneme.FeatBoundary),                                                                                                    // boundary lo=4194304
	1:  phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatLow, phoneme.FeatBack, phoneme.FeatSyllabic),                                             // AA lo=2179073
	2:  phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatLow, phoneme.FeatFront, phoneme.FeatSyllabic),                                            // AE lo=2146305
	3:  phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatMid, phoneme.FeatSyllabic),                                                               // AH lo=2105345
	4:  phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatStop, phoneme.FeatLow, phoneme.FeatSyllabic),                                             // AO lo=2113541
	5:  phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatLow, phoneme.FeatBack, phoneme.FeatDiphthong, phoneme.FeatSyllabic),                      // AW lo=2703361
	6:  phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatLow, phoneme.FeatFront, phoneme.FeatDiphthong, phoneme.FeatSyllabic),                     // AY lo=2670593
	7:  phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatStop, phoneme.FeatLabial),                                                                // B  lo=133
	8:  phoneme.FromBits(phoneme.FeatAffricate, phoneme.FeatPalatal),                                                                              // CH lo=528
	9:  phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatFricative),                                                                               // D  lo=9
	10: phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatFricative, phoneme.FeatAlveolar),                                                         // DH lo=265
	11: phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatMid, phoneme.FeatFront, phoneme.FeatSyllabic),                                            // EH lo=2138113
	12: phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatMid, phoneme.FeatRhotic, phoneme.FeatSyllabic),                                           // ER lo=3153921
	13: phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatMid, phoneme.FeatFront, phoneme.FeatTense, phoneme.FeatDiphthong, phoneme.FeatSyllabic),  // EY lo=2924545
	14: phoneme.FromBits(phoneme.FeatFricative, phoneme.FeatLabial),                                                                               // F  lo=136
	15: phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatStop),                                                                                    // G  lo=5
	16: phoneme.FromBits(phoneme.FeatFricative, phoneme.FeatGlottal),                                                                              // HH lo=2056
	17: phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatNasal, phoneme.FeatLabial),                                                               // M  lo=131
	18: phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatHigh, phoneme.FeatFront, phoneme.FeatTense, phoneme.FeatSyllabic),                        // IY lo=2396161
	19: phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatAffricate, phoneme.FeatPalatal),                                                          // JH lo=529
	20: phoneme.FromBits(phoneme.FeatStop, phoneme.FeatVelar),                                                                                     // K  lo=1028
	21: phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatApproximant, phoneme.FeatLateral, phoneme.FeatAlveolar),                                  // L  lo=353
	22: phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatMid, phoneme.FeatSyllabic),                                                               // IH lo=2105345 (same mask as AH in v0.1-en)
	23: phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatNasal, phoneme.FeatAlveolar),                                                             // N  lo=259
	24: phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatNasal, phoneme.FeatVelar),                                                                // NG lo=1027
	25: phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatMid, phoneme.FeatBack, phoneme.FeatRounded, phoneme.FeatDiphthong, phoneme.FeatSyllabic), // OW lo=2826241
	26: phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatMid, phoneme.FeatBack, phoneme.FeatRounded, phoneme.FeatDiphthong, phoneme.FeatSyllabic), // OY lo=2826241 (same as OW in v0.1-en)
	27: phoneme.FromBits(phoneme.FeatStop, phoneme.FeatLabial),                                                                                    // P  lo=132
	28: phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatApproximant, phoneme.FeatAlveolar),                                                       // R  lo=289
	29: phoneme.FromBits(phoneme.FeatFricative, phoneme.FeatAlveolar),                                                                             // S  lo=264
	30: phoneme.FromBits(phoneme.FeatFricative, phoneme.FeatPalatal),                                                                              // SH lo=520
	31: phoneme.FromBits(phoneme.FeatStop, phoneme.FeatAlveolar),                                                                                  // T  lo=260
	32: phoneme.FromBits(phoneme.FeatFricative, phoneme.FeatAlveolar),                                                                             // TH lo=264 (same as S in v0.1-en)
	33: phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatHigh, phoneme.FeatBack, phoneme.FeatRounded, phoneme.FeatSyllabic),                       // UH lo=2297857
	34: phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatHigh, phoneme.FeatBack, phoneme.FeatRounded, phoneme.FeatTense, phoneme.FeatSyllabic),    // UW lo=2560001
	35: phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatFricative, phoneme.FeatLabial),                                                           // V  lo=137
	36: phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatApproximant, phoneme.FeatLabial, phoneme.FeatVelar),                                      // W  lo=1185
	37: phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatApproximant, phoneme.FeatPalatal),                                                        // Y  lo=545
	38: phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatFricative, phoneme.FeatAlveolar),                                                         // Z  lo=265
	39: phoneme.FromBits(phoneme.FeatVoiced, phoneme.FeatFricative, phoneme.FeatPalatal),                                                          // ZH lo=521
}

// FeatureFor returns the FeatureMask for a neutral phoneme ID per the v0.1-en
// schema (ADR-0008). Unknown IDs return the zero mask. Note: ID 0 (boundary)
// returns its boundary mask (bit 22, lo=4194304), which is NOT the zero mask
// (the boundary-vs-zero disambiguation pinned by ADR-0007/ADR-0008).
func FeatureFor(id int) phoneme.FeatureMask {
	m, ok := phonemeFeatures[id]
	if !ok {
		return phoneme.Zero()
	}
	return m
}
