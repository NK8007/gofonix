# Changelog

All notable changes to this project are documented here. The format is based
on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.1-alpha] — 2026-06-06

### Changed
- **Release version metadata cleanup.** `ResultMetadata.GofonixVersion` now
  reports the Gofonix release/module version (`v0.3.1-alpha`) instead of the
  former `v0.1.0`. The result schema remains separately pinned as
  `SchemaVersion = "gofonix-result-v0.1"`; `FeatureSchemaVersion` remains
  `v0.1-en`. This is a metadata-semantics correction only — it is **not** a
  change to the JSON result schema.
- **Single source of truth for versions.** The release version, result-record
  schema version, and feature schema version are now defined once in
  `pkg/g2p/version.go` (`ReleaseVersion`, `schemaVersion`,
  `featureSchemaVersion`). `engine.go` no longer defines an independent
  `gofonixVersion = "v0.1.0"`, and the CLI no longer carries a duplicate
  `engineGofonixVersion`; `cmd/gofonix-cli` and the engine both report
  `g2p.ReleaseVersion`.
- **CLI `--version`** now reports `gofonix-cli v0.3.1-alpha` /
  `GofonixVersion: v0.3.1-alpha` / `SchemaVersion: gofonix-result-v0.1`
  consistently.

### Notes
- Golden corpus records were updated for the single field `gofonix_version`
  only; `schema_version`, `feature_schema_version`, `dictionary_id`,
  `dictionary_checksum`, `fallback_rules_version`, tokens, phonemes,
  alignment, and feature streams are unchanged.
- No changes to G2P logic, tokenization, fallback rules, ARPAbet mapping,
  alignment, `FeatureStream` semantics, the dictionary loader, or causal mode.

## [0.3.0-alpha] — 2026-06-03

### Added
- **ADR-0010 — Full CMUdict Loading Policy.** Frozen Phase 3 contract for the
  external full CMUdict v0.7b backend: opt-in via the `gofonix_full_dict`
  build tag, mandatory SHA-256 verification against a pinned digest, size
  band, and graceful degradation to the embedded mini-dict
  (`cmudict-mini-v0.1`) on any failure.
- **Full-dict loader.** `internal/dict/loader_full.go` (build tag
  `gofonix_full_dict`) plus a `!gofonix_full_dict` stub and sentinel-error
  surface (`FullDictFallbackReason`). The dictionary file itself is never
  committed (`.gitignore`).
- **Public API additions.** `g2p.Options.DictPath` (override resolution for
  the on-disk full CMUdict) and `g2p.ResultMetadata.FullDictAvailable` (the
  programmatic signal for "did the full CMUdict load successfully?"). New
  sentinel error `g2p.ErrDictionaryLoad`.
- **CLI — `gofonix-cli`.** Stdlib-only command-line entry point under
  `cmd/gofonix-cli`. Flags: `--language`, `--mode {batch|causal}`,
  `--input`, `--dict-path`, `--output {json|text}`, `--version`. CLI version
  reported as `v0.3.0-alpha`.
- **Library API polish.** Comprehensive godoc on every public symbol in
  `pkg/g2p` and `pkg/phoneme`; runnable `Example_basicUsage`,
  `Example_oovHandling`, and `Example_causalMode` in
  `pkg/g2p/example_test.go`.
- **Documentation.** Root `README.md`, `docs/architecture.md` (module map
  + ADR index), and this `CHANGELOG.md`.

### Notes
- `ResultMetadata.GofonixVersion` reports the Gofonix release/module version.
  The result schema is separately pinned as `SchemaVersion =
  "gofonix-result-v0.1"` (ADR-0001). (Superseded in 0.3.1-alpha: earlier
  pre-releases reported `GofonixVersion = "v0.1.0"`; this was a metadata
  inconsistency, corrected in 0.3.1-alpha so the field carries the release
  version while `SchemaVersion` remains the frozen result-record version.)
- Compliance files (`LICENSE`, `NOTICE`, `THIRD_PARTY_NOTICES`) are still
  absent from the repository root. They must be added before any non-alpha
  tag because the full-CMUdict opt-in places the user at the CMU/BSD license
  boundary (ADR-0010, License / Provenance).

## [0.2.0] — 2026-04-15

### Added
- **ADR-0009 — Deterministic OOV Fallback (`fallback-en-v0.2`).** Frozen
  Phase 2 contract for the deterministic English rule-based fallback that
  resolves `KindWord` tokens missed by the mini CMUdict.
- **OOV policy.** `ResultMetadata.OOVPolicy` is now
  `"rule-fallback-then-unknown"` and `ResultMetadata.FallbackRulesVersion`
  is `"fallback-en-v0.2"` whenever the English fallback is active. Tokens
  the fallback declines (and every non-word token) remain `SourceUnknown`.
- **New provenance value.** `g2p.SourceRuleFallback` is emitted for tokens
  resolved by the English fallback; `SourceDict` and `SourceUnknown`
  semantics are unchanged.

### Changed
- Amended ADR-0005 (OOV policy) per ADR-0009: from *unknown-only* to
  *rule-fallback-then-unknown* for English `KindWord` tokens. All other
  Phase 1 ADRs (ADR-0001, ADR-0003, ADR-0004, ADR-0006, ADR-0007, ADR-0008)
  remain authoritative and unchanged.

## [0.1.0] — 2026-02-01

### Added
- **Initial G2P engine.** Deterministic, dictionary-only English G2P core
  (`pkg/g2p`): `Engine`, `Options`, `Mode` (`ModeBatch`, `ModeOracle`,
  `ModeCausal` scaffold), `Process`, `Result`, `TokenResult`,
  `Pronunciation`, `ByteSpan`, `ResultMetadata`.
- **Mini CMUdict.** Embedded via `//go:embed`
  (`internal/dict/data/cmudict-mini-v0.1.dict`), surfaced as
  `DictionaryID = "cmudict-mini-v0.1"` with a SHA-256 checksum computed
  over the embedded bytes.
- **`FeatureMask`.** Opaque `struct{ bits [2]uint64 }` in `pkg/phoneme`
  with `IsZero`, `Lo`, `Hi`, `Equal`, `Has`. v0.1-en feature schema
  (ADR-0008): `FeatVoiced=0 … FeatBoundary=22`.
- **Byte-aligned `FeatureStream`.** Uniform-byte alignment per phoneme
  (ADR-0007); `len(FeatureStream) == len(Input)`.
- **Golden corpus.** `internal/golden` plus the schema documented in
  `docs/golden_jsonl_schema.md`.
- **Internal ARPAbet bridge.** `internal/lang/en/arpabet` — 39 symbols,
  neutral IDs 1–39. ARPAbet strings never leak into the public API
  (Principle 3).
- **ADRs ADR-0001, ADR-0003, ADR-0004, ADR-0005, ADR-0006, ADR-0007,
  ADR-0008** frozen for v0.1.

[Unreleased]: https://github.com/NK8007/gofonix/compare/v0.3.1-alpha...HEAD
[0.3.1-alpha]: https://github.com/NK8007/gofonix/releases/tag/v0.3.1-alpha
[0.3.0-alpha]: https://github.com/NK8007/gofonix/releases/tag/v0.3.0-alpha
[0.2.0]: https://github.com/NK8007/gofonix/releases/tag/v0.2.0
[0.1.0]: https://github.com/NK8007/gofonix/releases/tag/v0.1.0
