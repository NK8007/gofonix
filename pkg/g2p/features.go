package g2p

import (
	"github.com/NK8007/gofonix/internal/lang/en/arpabet"
	"github.com/NK8007/gofonix/pkg/phoneme"
)

// FeatureForPhoneme returns the v0.1-en FeatureMask for a neutral phoneme
// (ADR-0008). It is the public, language-neutral accessor for the same mask the
// engine projects into Result.FeatureStream, exposing only neutral phoneme IDs
// and opaque FeatureMasks — never RAW ARPAbet symbols (ADR-0002).
//
// The boundary / non-speech phoneme (ID 0) resolves to its FeatBoundary mask
// here, but that mask is never emitted into the FeatureStream in v0.1; the
// stream uses the zero mask for non-speech bytes (ADR-0007, ADR-0008).
func FeatureForPhoneme(p phoneme.Phoneme) phoneme.FeatureMask {
	return arpabet.FeatureFor(p.ID)
}
