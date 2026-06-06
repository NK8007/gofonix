# ADR-0007: Byte-Phoneme Alignment and FeatureStream Projection

## Status

Accepted.

## Context

CMUdict provides word-level pronunciations. It does not provide a grapheme-to-phoneme alignment for each input word. Gofonix nevertheless needs a byte-aligned `FeatureStream` for compression experiments.

The project therefore needs a deterministic projection from token-level pronunciations to byte-level feature masks. This projection must be reproducible, byte-based, and honest about its approximate nature.

## Decision

Gofonix uses uniform byte distribution to align phonemes to the bytes of a word token. It then projects each phoneme's feature mask onto the bytes assigned to that phoneme.

This alignment is an approximation. It is not a linguistic claim about exact grapheme-to-phoneme correspondence.

## ByteSpan semantics

All alignment spans are:

- zero-based;
- half-open;
- byte offsets into the original input;
- within the original word token span;
- independent of lookup normalization.

The invariant `0 <= Start <= End <= len(input)` must always hold.

## Uniform alignment algorithm

For a word token with byte span `[w0, w1)`:

- let `M = w1 - w0`, the number of bytes in the word token;
- let `N` be the number of phonemes in the pronunciation.

If `N == 0`, alignment is empty.

If `N >= 1`, the byte range `[w0, w1)` is partitioned into `N` contiguous spans. The partition covers the full word span exactly.

If `M` is not divisible by `N`, the remainder bytes are assigned to the leading phonemes. If `M < N`, some phonemes may receive zero-width spans.

Example:

```text
cat [0,3), phonemes K AE T
→ [0,1), [1,2), [2,3)
```

The same algorithm is used regardless of whether the pronunciation source is dictionary or deterministic fallback.

## Digraphs and many-to-one cases

Digraphs such as `th`, `sh`, `ch`, and `ng` receive no special treatment in this alignment algorithm. Uniform distribution may approximate them imperfectly, but it is deterministic and sufficient for the current byte-level feature projection.

Exact grapheme-to-phoneme alignment is future work.

## FeatureStream projection

`FeatureStream` has exactly `len(input)` entries, one per input byte.

For each phoneme with feature mask `Features[i]` and alignment span `[s, e)`, the projector writes that feature mask to every byte position in `FeatureStream[s:e]`.

Under the uniform alignment algorithm, spans are contiguous and non-overlapping, so no mask-combination step is needed. Each byte receives exactly one mask:

- a phoneme mask for bytes governed by a pronounced phoneme;
- the zero mask for bytes with no governing phoneme.

## Non-speech and unknown tokens

The following token categories receive zero masks in `FeatureStream`:

- whitespace;
- punctuation;
- numbers without pronunciation expansion;
- symbols;
- invalid/unknown tokens;
- word tokens whose pronunciation source is `SourceUnknown`.

Dictionary pronunciations and deterministic fallback pronunciations are projected using the normal phoneme alignment algorithm.

## Boundary phoneme versus zero mask

The boundary phoneme and the zero mask are distinct.

The boundary phoneme is a phoneme inventory item with its own non-zero feature mask. In the current `v0.1-en` schema, the boundary feature is bit 22, giving `lo = 4194304`.

The zero mask is `{lo:0, hi:0}` and means “no phoneme governs this byte.”

Current `FeatureStream` projection writes the zero mask for non-speech and unknown bytes. It does not emit the boundary phoneme into the stream.

Example:

```text
cat.
```

The bytes for `cat` receive masks from the pronunciation of `cat`. The byte for `.` receives the zero mask. That zero mask does not mean boundary phoneme ID 0 is present.

## Batch annotation versus causal prediction

`FeatureStream` has different interpretations by mode:

- in `ModeBatch` and `ModeOracle`, it is an annotation stream produced with full-input knowledge;
- in `ModeCausal`, it must be a predictor stream computable from prefix-only information.

Current `ModeCausal` is scaffold-only and emits zero masks for every byte. Non-zero causal feature projection requires a future prefix-safe implementation as specified in ADR-0003.

## Multi-byte UTF-8

Alignment is byte-based, not rune-based. A multi-byte UTF-8 code point occupies multiple `FeatureStream` entries.

Uniform byte partitioning may split a multi-byte code point across two phoneme spans. This is intentional for the current byte-level compression target. Rune-aware alignment may be considered in a future version, but it would be a different alignment policy.

## JSON representation

Feature masks in `FeatureStream` are serialized as:

```json
{"lo": 1028, "hi": 0}
```

This matches the golden JSONL representation and allows CLI JSON output to expose actual feature masks rather than empty objects.

## Alternatives considered

### Exact grapheme-to-phoneme alignment

Rejected for the current implementation. It would add substantial complexity and may introduce nondeterminism or model dependencies.

### Whole-word mask on every byte

Rejected. It would lose per-phoneme structure and make byte-level features less interpretable.

### OR-combine all phoneme masks over a word

Rejected. Uniform alignment already gives exactly one governing phoneme per byte. Combining masks would obscure provenance.

### Rune-aligned partition

Rejected for the current implementation. The compression target is byte-level, so byte-level alignment is the simpler and more direct contract.

## Consequences

- Every input byte has exactly one feature mask.
- The projection is deterministic and golden-testable.
- The alignment is explicitly approximate.
- Batch annotation and causal prediction semantics remain separate.
- JSON output and golden output use the same `{lo,hi}` mask shape.

