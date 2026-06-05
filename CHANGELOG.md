# Changelog

All notable changes to this project are documented here. The format is based
on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
- `ResultMetadata.GofonixVersion` is intentionally still reported as
  `"v0.1.0"` — the value is the frozen result-record version pinned by
  ADR-0001 and is decoupled from the module release tag.
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

[Unreleased]: https://github.com/NK8007/gofonix/compare/v0.3.0-alpha...HEAD
[0.3.0-alpha]: https://github.com/NK8007/gofonix/releases/tag/v0.3.0-alpha
[0.2.0]: https://github.com/NK8007/gofonix/releases/tag/v0.2.0
[0.1.0]: https://github.com/NK8007/gofonix/releases/tag/v0.1.0
