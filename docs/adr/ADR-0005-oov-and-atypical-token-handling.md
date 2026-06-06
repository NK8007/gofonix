# ADR-0005: OOV and Atypical Token Handling

## Status

Accepted for token classification and unknown-token semantics. Word-level OOV fallback policy is amended by ADR-0009.

## Context

Real text contains tokens that are absent from the dictionary or are not ordinary words. Gofonix must handle these cases deterministically so that golden tests and compression experiments are reproducible.

The project also needs a clear distinction between:

- dictionary pronunciations;
- deterministic fallback pronunciations;
- tokens with no pronunciation.

The unknown-only behavior remains a useful baseline for ablation and for unsupported tokens. Current Gofonix uses the amended default policy from ADR-0009: `rule-fallback-then-unknown`.

## Decision

Gofonix classifies each token by token kind and assigns pronunciation source explicitly.

The current OOV policy is:

```text
dictionary lookup → deterministic rule fallback → unknown
```

That policy is reported as:

```text
OOVPolicy = "rule-fallback-then-unknown"
```

The fallback rule set is versioned separately through `FallbackRulesVersion`.

## Source behavior

### Dictionary hit

If a word token has a usable dictionary pronunciation:

- `SourceDict` is emitted;
- phonemes, features, and alignment are populated;
- token bytes receive projected feature masks in `FeatureStream`.

### Rule fallback

If a word token is not resolved by dictionary lookup but is supported by deterministic fallback:

- `SourceRuleFallback` is emitted;
- fallback phonemes and features are populated;
- alignment and `FeatureStream` projection follow the same rules as dictionary pronunciations.

The fallback policy itself is specified by ADR-0009.

### Unknown

If no pronunciation is produced:

- `SourceUnknown` is emitted;
- phoneme list is empty;
- feature list is empty;
- alignment is empty;
- the token's byte positions receive the zero mask in `FeatureStream`.

This applies to unsupported OOV words and to non-speech tokens.

## Atypical tokens

### Whitespace

Whitespace tokens do not produce pronunciations. Their bytes receive the zero mask.

### Punctuation

Punctuation tokens do not produce pronunciations. Their bytes receive the zero mask.

### Numbers

Numbers are tokenized deterministically but are not expanded into words in the current implementation. They normally produce `SourceUnknown` and zero feature masks.

### Symbols

Symbols do not produce pronunciations. Their bytes receive the zero mask.

### Invalid or unknown bytes

Invalid UTF-8 and otherwise unclassified bytes become unknown tokens. They preserve their original byte spans and receive zero feature masks.

### Mixed alphanumeric tokens

Mixed alphanumeric runs containing at least one letter are tokenized as words. They may be resolved by dictionary lookup, by deterministic fallback if supported, or by `SourceUnknown`.

Examples include:

- `MP3`;
- `H2O`;
- `3rd`;
- `21st`.

The tokenizer classifies these as word tokens, not number tokens.

### Acronyms and unusual words

Acronyms and unusual letter strings are word tokens. Their pronunciation source depends on dictionary and fallback behavior.

For example, a token such as `NASA` may be handled by deterministic fallback if it is absent from the dictionary and the fallback rules support it.

## What Gofonix does not do in the current policy

The current implementation does not perform:

- neural or statistical G2P;
- number-to-words expansion;
- named-entity recognition;
- abbreviation expansion;
- probabilistic source confidence scoring.

All behavior is deterministic and source-labeled.

## Consequences

- OOV and non-word behavior is reproducible.
- Fallback-derived pronunciations are distinguishable from dictionary pronunciations.
- Unknown tokens remain explicit rather than silently inventing unsupported phonology.
- The unknown-only baseline remains conceptually useful for ablation, but the current default policy is the amended rule-fallback policy from ADR-0009.
