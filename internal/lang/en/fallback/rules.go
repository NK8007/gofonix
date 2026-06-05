// Package fallback implements the deterministic, rule-based English OOV
// grapheme-to-phoneme fallback frozen as `fallback-en-v0.2` (ADR-0009).
//
// It is a pure, hermetic, standard-library-only package: a static
// grapheme -> ARPAbet-symbol-sequence table plus a deterministic greedy
// longest-match left-to-right scanner. It performs no runtime I/O, no network
// access, and no subprocess/CGO. Same input -> identical output, always.
//
// The right-hand-side symbols are internal ARPAbet-like strings drawn from the
// 39-symbol v0.1-en inventory. They are intended to be handed to the existing
// internal/lang/en/arpabet bridge; this package never exposes ARPAbet through
// the public pkg/g2p API. This package carries no API stability guarantee.
//
// Scope note (ADR-0009, Phase 2 / Slice 2): this slice provides only the rule
// table, normalization, and the longest-match matcher. It does NOT integrate
// with pkg/g2p Process(), does NOT emit SourceRuleFallback, and does NOT add
// ResultMetadata.FallbackRulesVersion. Those land in a later slice.
package fallback

// RulesVersion identifies the frozen fallback rule set defined by ADR-0009.
// Changing the rule table requires bumping this version string.
const RulesVersion = "fallback-en-v0.2"

// arpabetInventory is the 39-symbol v0.1-en ARPAbet inventory (ADR-0008).
// Every right-hand-side symbol in the rule table must be a member of this set.
// It mirrors internal/lang/en/arpabet/symbols.go but is duplicated here as a
// validation-only set so that validateRules has a local, hermetic source of
// truth and the fallback package does not depend on the bridge for validation.
var arpabetInventory = map[string]struct{}{
	"AA": {}, "AE": {}, "AH": {}, "AO": {}, "AW": {}, "AY": {},
	"B": {}, "CH": {}, "D": {}, "DH": {},
	"EH": {}, "ER": {}, "EY": {},
	"F": {}, "G": {}, "HH": {},
	"IH": {}, "IY": {}, "JH": {}, "K": {},
	"L": {}, "M": {}, "N": {}, "NG": {},
	"OW": {}, "OY": {}, "P": {}, "R": {},
	"S": {}, "SH": {}, "T": {}, "TH": {},
	"UH": {}, "UW": {}, "V": {}, "W": {},
	"Y": {}, "Z": {}, "ZH": {},
}

// rule is one entry of the frozen fallback table: a grapheme (1+ ASCII letters)
// mapped to an ordered sequence of internal ARPAbet symbols. Note that a single
// grapheme may emit more than one symbol (e.g. `x` -> {"K","S"}).
type rule struct {
	grapheme string
	symbols  []string
}

// rules is the complete, authoritative, frozen `fallback-en-v0.2` rule table,
// transcribed verbatim from ADR-0009. The order here is for human readability
// only; the matcher's behaviour is fully determined by grapheme length plus
// alphabetical tie-break, independent of slice/map iteration order.
//
// Duplicate left-hand-side keys are invalid (enforced by validateRules).
var rules = []rule{
	// Digraphs (multi-letter consonant clusters).
	{"sh", []string{"SH"}},
	{"ch", []string{"CH"}},
	{"th", []string{"DH"}},
	{"ph", []string{"F"}},
	{"wh", []string{"W"}},
	{"ng", []string{"NG"}},
	{"ck", []string{"K"}},

	// Common vowel patterns.
	{"ee", []string{"IY"}},
	{"ea", []string{"IY"}},
	{"oo", []string{"UW"}},
	{"ai", []string{"EY"}},
	{"ay", []string{"EY"}},
	{"ou", []string{"AW"}},
	{"ow", []string{"OW"}},
	{"oi", []string{"OY"}},
	{"oy", []string{"OY"}},

	// Single consonants.
	{"b", []string{"B"}},
	{"c", []string{"K"}},
	{"d", []string{"D"}},
	{"f", []string{"F"}},
	{"g", []string{"G"}},
	{"h", []string{"HH"}},
	{"j", []string{"JH"}},
	{"k", []string{"K"}},
	{"l", []string{"L"}},
	{"m", []string{"M"}},
	{"n", []string{"N"}},
	{"p", []string{"P"}},
	{"q", []string{"K"}},
	{"r", []string{"R"}},
	{"s", []string{"S"}},
	{"t", []string{"T"}},
	{"v", []string{"V"}},
	{"w", []string{"W"}},
	{"x", []string{"K", "S"}}, // a single grapheme emitting TWO symbols
	{"y", []string{"Y"}},
	{"z", []string{"Z"}},

	// Simple vowels (single-letter defaults).
	{"a", []string{"AE"}},
	{"e", []string{"EH"}},
	{"i", []string{"IH"}},
	{"o", []string{"AO"}},
	{"u", []string{"AH"}},
}

// ruleIndex maps a grapheme key to its ARPAbet symbol sequence, built once from
// the frozen `rules` slice for O(1) longest-match probing. maxGraphemeLen is the
// length of the longest grapheme key in the table.
var (
	ruleIndex     map[string][]string
	maxGraphemeLen int
)

func init() {
	ruleIndex = make(map[string][]string, len(rules))
	for _, r := range rules {
		ruleIndex[r.grapheme] = r.symbols
		if len(r.grapheme) > maxGraphemeLen {
			maxGraphemeLen = len(r.grapheme)
		}
	}
}
