//go:build !gofonix_full_dict

// Stub LoadFull for the default build (no `gofonix_full_dict` build tag).
// The function returns ErrFullDictNotBuilt immediately so the G2P engine
// emits no warning (ADR-0010, Fallback Policy: default-build path is silent)
// and falls back to the embedded mini-dict.
//
// The signature MUST match loader_full.go byte-for-byte so the engine code
// compiles identically under both build tags.

package dict

// LoadFull is the stub variant: this binary was built without
// `-tags gofonix_full_dict`, so the full CMUdict loader is intentionally
// absent. Callers receive ErrFullDictNotBuilt and degrade gracefully to the
// embedded mini-dict (ADR-0010, Decision §7).
func LoadFull(override string) (*Dict, error) {
	// The override argument is accepted (and ignored) so the stub and the
	// real loader share the same signature; callers may always pass a path
	// without knowing which variant is compiled in.
	_ = override
	return nil, ErrFullDictNotBuilt
}
