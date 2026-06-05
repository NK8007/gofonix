package g2p

// Centralised version identifiers for the public Gofonix package.
//
// This file is the single source of truth for the three version strings that
// downstream code (the engine, the CLI, tests, and documentation) must agree
// on. Keeping them here avoids the previous situation where independent string
// literals (notably "v0.1.0") were duplicated across engine.go and the CLI and
// could silently drift apart.
//
// Semantic distinction (do not conflate these):
//
//   - ReleaseVersion  — the Gofonix release/module version. This is what the
//     engine reports as ResultMetadata.GofonixVersion and what the CLI prints
//     for --version. It tracks the module/tag (e.g. v0.3.1-alpha).
//   - schemaVersion   — the frozen result-record schema version (the on-disk
//     JSONL contract). It is intentionally independent of the release version
//     and only changes when the result-record format itself changes.
//   - featureSchemaVersion — the phonological feature schema version (ADR-0008),
//     also independent of the release version.
//
// Other reproducibility identifiers (DictionaryID, DictionaryChecksum,
// OOVPolicy, FallbackRulesVersion, NormalizerVersion, TokenizerVersion) remain
// defined alongside the engine logic that owns them; they are deliberately not
// centralised here because they are not "the Gofonix version".
const (
	// ReleaseVersion is the Gofonix release/module version. It is reported as
	// ResultMetadata.GofonixVersion and is the value the CLI prints for
	// --version. It is distinct from schemaVersion and featureSchemaVersion.
	ReleaseVersion = "v0.3.1-alpha"

	// schemaVersion is the frozen result-record schema version recorded as
	// Result.SchemaVersion. It is the on-disk JSONL contract and changes only
	// when the result-record format changes — never merely because the release
	// version changes.
	schemaVersion = "gofonix-result-v0.1"

	// featureSchemaVersion is the phonological feature schema version (ADR-0008)
	// recorded as ResultMetadata.FeatureSchemaVersion. It is independent of the
	// release version.
	featureSchemaVersion = "v0.1-en"
)
