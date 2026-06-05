// Package g2p is the Gofonix Grapheme-to-Phoneme engine: options, modes, the
// Engine, and the Result types (ADR-0001).
//
// The public API is batch-only: the full input string is supplied in a single
// Process call. ModeBatch, ModeOracle, and ModeCausal are accepted, but a
// separately-callable streaming Causal API is deferred (ADR-0001, ADR-0003).
//
// In ModeBatch/ModeOracle a word token is resolved against the active CMUdict
// first (ADR-0004, ADR-0010): a hit yields SourceDict; a miss is then attempted
// through the deterministic English rule fallback (fallback-en-v0.2, ADR-0009)
// and yields SourceRuleFallback when the fallback resolves, or SourceUnknown
// otherwise. Every resolved token carries a per-phoneme uniform byte Alignment
// (one ByteSpan per phoneme, ADR-0007) and Variant 0; each phoneme's v0.1-en
// FeatureMask is projected over its alignment span into Result.FeatureStream.
// The FeatureStream always has exactly len(input) entries.
//
// ModeCausal is a scaffold (ADR-0003): every token is SourceUnknown and the
// FeatureStream stays all-zero. A real streaming Causal simulator is deferred.
//
// The active dictionary backend is selected at build time (ADR-0010): default
// builds ship only the embedded mini-dict ("cmudict-mini-v0.1"); under
// -tags gofonix_full_dict the engine attempts to load the full CMUdict v0.7b
// from disk and falls back gracefully to the mini-dict if the file is missing,
// malformed, or fails its SHA-256 check. ResultMetadata.FullDictAvailable and
// ResultMetadata.DictionaryID are the programmatic signals for which backend
// was used.
package g2p
