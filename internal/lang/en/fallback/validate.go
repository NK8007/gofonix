package fallback

import (
	"fmt"
	"sort"
)

// validateRules checks the integrity of the frozen `fallback-en-v0.2` rule
// table (ADR-0009 "Rule Representation"). It is invoked from tests (and is
// cheap enough to run at init-time if desired). It enforces:
//
//   - no empty grapheme keys;
//   - no duplicate left-hand-side grapheme keys;
//   - every right-hand-side symbol is a member of the 39-symbol v0.1-en
//     ARPAbet inventory;
//   - single-letter rules exist for all 26 ASCII letters a-z;
//   - the maximum grapheme length is computable and positive.
//
// It returns nil when the table is valid, otherwise a descriptive error.
func validateRules() error {
	seen := make(map[string]struct{}, len(rules))
	computedMax := 0

	for _, r := range rules {
		if r.grapheme == "" {
			return fmt.Errorf("fallback: empty grapheme key in rule table")
		}
		if _, dup := seen[r.grapheme]; dup {
			return fmt.Errorf("fallback: duplicate grapheme key %q", r.grapheme)
		}
		seen[r.grapheme] = struct{}{}

		if len(r.symbols) == 0 {
			return fmt.Errorf("fallback: grapheme %q has no ARPAbet symbols", r.grapheme)
		}
		for _, sym := range r.symbols {
			if _, ok := arpabetInventory[sym]; !ok {
				return fmt.Errorf("fallback: grapheme %q maps to symbol %q outside the v0.1-en inventory", r.grapheme, sym)
			}
		}

		if len(r.grapheme) > computedMax {
			computedMax = len(r.grapheme)
		}
	}

	// Single-letter rules must exist for every ASCII letter a-z so the
	// longest-match scanner can always make progress over letter-only input.
	var missing []string
	for c := byte('a'); c <= 'z'; c++ {
		if _, ok := seen[string(c)]; !ok {
			missing = append(missing, string(c))
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("fallback: missing single-letter rule(s) for %v", missing)
	}

	if computedMax <= 0 {
		return fmt.Errorf("fallback: max grapheme length must be positive, got %d", computedMax)
	}
	// Cross-check the init-computed maxGraphemeLen against a fresh computation.
	if computedMax != maxGraphemeLen {
		return fmt.Errorf("fallback: maxGraphemeLen mismatch: computed %d, cached %d", computedMax, maxGraphemeLen)
	}

	return nil
}
