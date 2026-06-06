# ADR-0004: CMUdict Usage Strategy

## Status

Accepted. Full-dictionary loading policy is extended by ADR-0010.

## Context

English pronunciation data in Gofonix comes from the CMU Pronouncing Dictionary. CMUdict is ARPAbet-based, while the public Gofonix model uses neutral phoneme IDs and feature masks.

The project needs dictionary use to be:

- reproducible;
- license-compliant;
- deterministic;
- small enough for normal tests;
- extensible to a full dictionary backend for larger experiments.

## Decision

Gofonix uses a committed mini CMUdict fixture as the default dictionary and supports an optional full CMUdict backend under a separate policy.

The mini fixture is embedded into the Go binary and is sufficient for unit tests, golden tests, smoke tests, and small examples. It has its own stable dictionary ID and checksum.

The full CMUdict is not treated as a normal always-bundled runtime dependency. Its loading and fallback policy are specified by ADR-0010.

## Mini dictionary

The default dictionary is:

- embedded with `go:embed`;
- identified as `cmudict-mini-v0.1`;
- checksummed with SHA-256;
- reported through `Result.Metadata.DictionaryID`;
- reported through `Result.Metadata.DictionaryChecksum`.

The mini fixture is intentionally small. It covers golden-test vocabulary, smoke-test vocabulary, and selected edge cases.

## License and attribution

CMUdict license and attribution material must be preserved in repository documentation. The project should retain:

- the upstream license text or an equivalent license notice;
- third-party notices where appropriate;
- dictionary identity and checksum in reproducibility metadata.

Experiments using a full dictionary should record the upstream source, retrieval date, checksum, and any local modifications.

## Lookup policy

Dictionary lookup uses a normalized lookup key derived from the original token. The original token and byte span remain unchanged.

For a dictionary hit:

- the selected pronunciation is mapped from ARPAbet to neutral phoneme IDs;
- stress digits are stripped before phoneme mapping;
- `SourceDict` is recorded;
- `Variant` is `0` in the current public result model.

For a dictionary miss:

- the current OOV policy is applied;
- deterministic fallback may produce a pronunciation;
- otherwise the token becomes `SourceUnknown`.

## Alternate pronunciations

CMUdict can contain alternate pronunciations such as `WORD(1)` and `WORD(2)`. Current Gofonix keeps pronunciation choice deterministic by selecting the base unnumbered entry as the canonical dictionary pronunciation.

Non-base variants are not exposed as selectable options in the current public API. Future versions may add controlled variant selection, but doing so would require explicit documentation and tests.

## Stress handling

ARPAbet vowels may include stress digits:

- `0`: unstressed;
- `1`: primary stress;
- `2`: secondary stress.

Current Gofonix strips stress digits before mapping to neutral phoneme IDs. Stress is not represented in `v0.1-en` feature masks. Adding stress as a feature would require a new feature schema version.

## Parsing policy

The dictionary parser should be deterministic:

- comment lines are ignored;
- entries are parsed as a word followed by one or more ARPAbet tokens;
- lookup keys are normalized consistently with ADR-0006;
- unsupported or invalid pronunciation data must not produce nondeterministic output;
- parser behavior must be covered by tests.

## Full dictionary policy

The default mini dictionary is always available. The optional full dictionary is governed by ADR-0010.

When full dictionary support is enabled but the full dictionary is unavailable or invalid, Gofonix must fall back safely to the mini dictionary rather than making normal library use depend on local data availability.

## Consequences

- Default builds are small, deterministic, and hermetic.
- Tests do not depend on network access or external dictionary files.
- Larger experiments can opt into a full dictionary path without changing the public result model.
- Dictionary identity is visible in result metadata, so outputs from mini and full backends are distinguishable.
- Stress loss is explicit and can be revisited through a future schema version.

