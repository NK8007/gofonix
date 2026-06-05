package arpabet

import (
	"strings"

	"github.com/NK8007/gofonix/pkg/phoneme"
)

// MapSymbol maps a single (possibly stressed) ARPAbet symbol to a neutral
// phoneme.Phoneme. The symbol is upper-cased and stress-stripped before lookup.
// Returns (phoneme.Phoneme{}, false) for unknown symbols. The returned Phoneme
// carries only its neutral ID; the ARPAbet string never leaves this package.
func MapSymbol(sym string) (phoneme.Phoneme, bool) {
	id, ok := symbolToID[StripStress(strings.ToUpper(sym))]
	if !ok {
		return phoneme.Phoneme{}, false
	}
	return phoneme.Phoneme{ID: id}, true
}

// MapSequence maps a slice of ARPAbet symbols to neutral Phonemes. If any symbol
// is unknown, MapSequence returns (nil, false): in v0.1 any unknown symbol in a
// word's pronunciation causes the whole word to fall back to SourceUnknown
// (ADR-0005). On success it returns the neutral phoneme sequence and true.
func MapSequence(syms []string) ([]phoneme.Phoneme, bool) {
	out := make([]phoneme.Phoneme, 0, len(syms))
	for _, s := range syms {
		p, ok := MapSymbol(s)
		if !ok {
			return nil, false
		}
		out = append(out, p)
	}
	return out, true
}
