# ADR-0002: Phoneme and Feature Representation

## Status

Accepted.

## Context

Gofonix needs phonological features, not only phoneme labels. The implementation currently targets English, but the public model should not be tied to ARPAbet or any other English-specific encoding.

CMUdict provides English pronunciations in ARPAbet. ARPAbet is useful as an internal input format, but it is not appropriate as the public identity model because it would make later language support harder and would expose implementation details in persisted outputs.

The project therefore needs:

- a neutral phoneme type;
- a compact feature representation;
- an explicitly versioned feature schema;
- stable JSON representation for feature masks in CLI and artifact output.

## Decision

The public model uses language-neutral `phoneme.Phoneme` values. English ARPAbet symbols are mapped to neutral phoneme IDs inside the English internal module.

IPA is treated as a human-readable rendering for debugging and documentation. It is not the canonical internal identity.

Phonological features are represented as an opaque `FeatureMask`. Callers may inspect masks through public methods and JSON output, but they should not depend on internal bit storage.

The current `FeatureMask` has two 64-bit words, giving 128 available feature bits. The active English feature schema uses only the low word. The feature inventory and bit assignments are defined by ADR-0008.

## Representation pipeline

```text
CMUdict ARPAbet token
        ↓
internal English ARPAbet mapping
        ↓
neutral phoneme.Phoneme
        ↓
phoneme.FeatureMask under feature schema v0.1-en
```

The ARPAbet mapping and the neutral-to-feature mapping are deterministic tables. Changing either table in a way that affects output requires a versioned documentation and test update.

## Phoneme identity

A public phoneme consists of:

- a stable neutral ID;
- a kind, such as consonant, vowel, or boundary.

The current English inventory is defined in ADR-0008. IDs are assigned so that current English phonemes and future language modules can coexist without depending on ARPAbet as the public namespace.

## Feature masks

`FeatureMask` is an opaque fixed-width set of phonological features.

The design rules are:

- masks are meaningful only together with their feature schema version;
- feature bit positions are never silently reused;
- any change to feature membership or bit assignment requires a new schema version;
- callers should not construct arbitrary masks unless the public API explicitly supports that use case;
- the zero mask means “no feature mask applies here,” not “boundary phoneme.”

The distinction between boundary phoneme and zero mask is specified in ADR-0007 and ADR-0008.

## JSON representation

Although `FeatureMask` is opaque in Go, it has a stable JSON representation for result output:

```json
{"lo": 2056, "hi": 0}
```

The fields represent the low and high 64-bit words of the mask. This format is used by CLI JSON output, `FeatureStream`, token-level features, and golden JSONL records.

The JSON representation is intentionally simple and schema-qualified by metadata rather than self-describing. Consumers must interpret the numbers under the accompanying `FeatureSchemaVersion`.

## Alternatives considered

### ARPAbet as the public model

Rejected. ARPAbet is English-specific and would make the public API less suitable for future multilingual extensions.

### IPA as the public model

Rejected. IPA is useful for human display, but string comparison and Unicode normalization make it unsuitable as the canonical internal identity.

### Raw integer feature masks

Rejected. A raw `uint64` alias would expose the bit layout as the whole abstraction and would make schema evolution harder. The implementation can still serialize low and high words while preserving an opaque Go type.

### Map-based feature sets

Rejected. Maps are larger, less compact, and do not have a deterministic iteration order. A fixed bitset is compact and deterministic.

## Consequences

- Public code does not need to know about ARPAbet.
- Feature masks are compact enough for byte-level `FeatureStream` use.
- Persisted masks remain interpretable because the feature schema version is recorded in result metadata and golden records.
- JSON output is now explicit and stable instead of exposing empty objects for unexported Go fields.
- Future schema changes can be made deliberately through new schema versions rather than silent bit reinterpretation.

