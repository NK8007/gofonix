# ADR-0003: Batch, Oracle, and Causal Semantics

## Status

Accepted.

## Context

Gofonix is intended to support byte-level predictive compression experiments. In a fair predictive setting, the model at byte position `i` may use only information available before byte `i`. Any feature that depends on byte `i` or later bytes is information leakage.

At the same time, full-input annotation is still useful. Batch processing can provide diagnostic features, upper-bound signals, golden test fixtures, and analysis outputs. The project therefore needs explicit mode semantics so that annotation results are not mistaken for fair causal compression results.

## Decision

Gofonix defines three modes:

- `ModeBatch`: full input text is available;
- `ModeOracle`: future-aware analysis mode reserved for upper-bound experiments;
- `ModeCausal`: no-look-ahead mode for fair byte-by-byte predictive compression.

Only `ModeCausal` is admissible for headline causal compression results.

## No-look-ahead definition

A feature is causally available at byte position `i` if and only if it can be computed exclusively from bytes before `i`.

For the same decoded prefix, causal output must be identical regardless of future bytes. Formally, for any prefix `P` and future continuations `A` and `B`, the causal feature for the next byte after `P` must be the same for `P + A` and `P + B`.

This invariant is the core leakage boundary.

## Mode semantics

### ModeBatch

`ModeBatch` may inspect the full input string. Tokenization, dictionary lookup, fallback, alignment, and feature projection may all depend on bytes at or after the byte currently being annotated.

`ModeBatch` is suitable for:

- deterministic annotation;
- golden tests;
- offline analysis;
- upper-bound or ablation experiments clearly labeled as non-causal.

It is not a fair streaming predictor mode.

### ModeOracle

`ModeOracle` represents a future-aware upper-bound regime. It may use full word or surrounding context. It exists to make upper-bound assumptions explicit rather than hidden inside batch behavior.

Oracle results must not be reported as fair byte-by-byte compression results.

### ModeCausal

`ModeCausal` is the fair predictive mode. A non-trivial causal implementation must compute the feature for each byte using only the decoded prefix before that byte.

In the current implementation, `ModeCausal` is scaffold-only. It accepts input and returns a structurally valid result, but every `FeatureStream` entry is the zero mask. This keeps the API and leakage-test plumbing present while avoiding a false claim of causal phonological inference.

## Future causal implementation contract

A future causal feature function would have the following semantic shape:

```go
func CausalFeatures(decodedPrefix []byte) (phoneme.FeatureMask, error)
```

The function must be pure with respect to `decodedPrefix`. It must not depend on:

- the full input;
- bytes at or after the target position;
- global mutable caches containing future-derived information;
- call history;
- wall-clock time;
- randomness.

## Test requirements

A non-trivial causal implementation must be tested for:

- determinism for repeated calls with the same prefix;
- prefix stability under arbitrary future continuations;
- invariance of earlier positions when later bytes are mutated;
- absence of hidden full-input state in tokenization, lookup, fallback, and projection.

The current scaffold trivially satisfies leakage tests because it always emits zero masks. When non-zero causal features are implemented, those tests must be strengthened with non-trivial expected behavior.

## Consequences

- Batch annotation and causal prediction are not conflated.
- Current Gofonix results are honest about causal limitations.
- Future causal work has a precise semantic target.
- Compression experiments can label batch/oracle numbers as upper bounds and causal numbers as fair only when causal features are genuinely prefix-derived.

