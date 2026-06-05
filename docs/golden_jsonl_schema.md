# Golden JSONL schema (`gofonix-result-v0.1`)

This document defines the **golden record** format used by Gofonix's
conformance and regression tests (Phase 1 / Slice 5). A golden file is a
[JSON Lines](https://jsonlines.org/) (`.jsonl`) file: one independent,
self-describing JSON object per line, UTF-8 encoded, each line terminated by a
single `\n`.

Each line is a frozen snapshot of `g2p.Engine.Process(input)` for a fixed
`(input, mode)` pair. The golden comparison tests deserialize each line, re-run
`Process`, re-serialize the live result, and require an **exact** match. The
serializer/deserializer lives in `internal/golden`.

The schema is intentionally **phoneme-neutral**: it carries only neutral
integer phoneme IDs (ADR-0008) and opaque `{lo,hi}` feature words (ADR-0002). It
**never** carries RAW ARPAbet symbols — ARPAbet is an internal implementation
detail and must not leak into any serialized artifact.

## Top-level object

| Field                    | Type            | Description |
|--------------------------|-----------------|-------------|
| `schema_version`         | string          | Result-record schema version. Always `"gofonix-result-v0.1"` in v0.1. |
| `gofonix_version`        | string          | Gofonix release/module version, e.g. `"v0.3.1-alpha"`. Distinct from `schema_version`, which pins the result-record format. |
| `dictionary_id`          | string          | Dictionary identity, e.g. `"cmudict-mini-v0.1"` (ADR-0004). |
| `dictionary_checksum`    | string          | Hex SHA-256 of the dictionary file used (ADR-0004). |
| `feature_schema_version` | string          | Phonological feature schema version, e.g. `"v0.1-en"` (ADR-0008). |
| `normalizer_version`     | string          | Normalizer version (ADR-0006). |
| `tokenizer_version`      | string          | Tokenizer version (ADR-0006). |
| `oov_policy`             | string          | Out-of-vocabulary policy, e.g. `"unknown-only"` (ADR-0005). |
| `language`               | string          | Resolved language tag (`"en"` in v0.1). |
| `mode`                   | string          | Analysis mode: `"batch"`, `"oracle"`, or `"causal"` (ADR-0003). |
| `input`                  | string          | The exact input string passed to `Process`. |
| `feature_stream`         | array of `Mask` | Byte-aligned feature stream; **exactly `len(input)` entries** (ADR-0007). |
| `tokens`                 | array of `Token`| Deterministic token sequence. |

> The `mode` field uses stable string spellings (`batch`/`oracle`/`causal`)
> that are decoupled from the Go `Mode` enum's integer values, so reordering the
> enum can never silently rewrite the golden corpus.

## `Mask`

A `Mask` is the JSON form of an opaque `phoneme.FeatureMask` — the two raw
64-bit words.

| Field | Type             | Description |
|-------|------------------|-------------|
| `lo`  | uint64 (number)  | Low 64 feature bits. |
| `hi`  | uint64 (number)  | High 64 bits. Reserved; always `0` under `v0.1-en` (ADR-0008). |

- A **non-speech byte** (whitespace, punctuation, number, symbol, unknown, OOV)
  is the **zero mask** `{ "lo": 0, "hi": 0 }`.
- The **boundary phoneme** (ID 0) mask `{ "lo": 4194304, "hi": 0 }` is **never**
  emitted into `feature_stream` in v0.1; non-speech is the zero mask instead
  (ADR-0007, ADR-0008).
- In `causal` mode the entire `feature_stream` is all-zero (the causal scaffold
  performs no projection in v0.1; ADR-0003).

## `Span`

A half-open `[start, end)` range of **byte** offsets into `input`.

| Field   | Type | Description |
|---------|------|-------------|
| `start` | int  | Inclusive byte offset. |
| `end`   | int  | Exclusive byte offset. |

Invariant: `0 <= start <= end <= len(input)`. Spans are byte offsets, not rune
offsets, and may split a multi-byte code point (ADR-0007). An empty span
`[i, i)` is valid.

## `Phoneme`

| Field | Type | Description |
|-------|------|-------------|
| `id`  | int  | Neutral phoneme ID. IDs 1–39 are the v0.1-en inventory; ID 0 is the universal boundary symbol (ADR-0008). |

The debug-only IPA rendering is deliberately **not** serialized.

## `Token`

| Field       | Type               | Description |
|-------------|--------------------|-------------|
| `token`     | string             | Exact surface substring `input[span.start:span.end]`. |
| `span`      | `Span`             | Original-byte range of the token. |
| `kind`      | string             | One of `word`, `whitespace`, `punctuation`, `number`, `symbol`, `unknown`. |
| `source`    | string             | Provenance: `dict`, `rule_fallback` (reserved, never emitted in v0.1), or `unknown`. |
| `variant`   | int                | Selected pronunciation variant; always `0` in v0.1 (ADR-0004). |
| `phonemes`  | array of `Phoneme` | Phoneme sequence; empty when `source == "unknown"`. |
| `alignment` | array of `Span`    | One byte span per phoneme via uniform byte distribution (ADR-0007); `len(alignment) == len(phonemes)`. Empty when `source == "unknown"`. |
| `features`  | array of `Mask`    | One feature mask per phoneme (the `v0.1-en` mask for each ID). `len(features) == len(phonemes)`. Convenience: lets a record be checked without re-deriving the table. |

ARPAbet is **not** required and **not** present in golden output.

## Invariants checked by the comparison tests

1. `len(feature_stream) == len(input)` (byte count, not rune count).
2. Every token span is a byte range with `0 <= start <= end <= len(input)`.
3. Non-speech / OOV tokens carry `source == "unknown"`, empty `phonemes`, empty
   `alignment`, and project only the zero mask over their bytes.
4. `mode == "causal"` ⇒ all-zero `feature_stream` **and** every token is
   `source == "unknown"` with empty `phonemes`/`alignment`.
5. The boundary mask `{lo:4194304}` never appears in `feature_stream`.
6. A dict-hit word's `alignment` is a contiguous, non-overlapping partition of
   its byte span, one span per phoneme.

## Reference values (ADR-0008, mini CMUdict `cmudict-mini-v0.1`)

| Word  | Phoneme IDs (after stress strip) | Per-phoneme `lo` |
|-------|----------------------------------|------------------|
| `cat` | K(20) AE(2) T(31)                | 1028, 2146305, 260 |
| `dog` | D(9) AO(4) G(15)                 | 9, 2113541, 5 |
| `it`  | IH(22) T(31)                     | 2105345, 260 |
| `the` | DH(10) IH(22)                    | 265, 2105345 |
| `a`   | AH(3)                            | 2105345 |

Boundary ID 0 ⇒ `lo = 4194304` (never emitted to the stream).

## Updating goldens

Golden files are **never** auto-rewritten by a normal test run. Regenerate them
intentionally with the update flag and review the diff before committing — see
[`testdata/golden/README.md`](../testdata/golden/README.md).

```sh
go test ./pkg/g2p/ -run TestGolden -update
# or:
GOFONIX_UPDATE_GOLDEN=1 go test ./pkg/g2p/ -run TestGolden
```
