# ADR-0006: Tokenization and Normalization

## Status

Accepted.

## Context

Gofonix needs deterministic tokenization because every downstream result depends on token spans. The tokenizer feeds dictionary lookup, fallback, phoneme alignment, and byte-level feature projection.

The central requirement is that spans always refer to original input bytes. Normalization may be used to derive lookup keys, but it must not alter the canonical byte positions stored in results.

## Decision

Gofonix uses a deterministic left-to-right tokenizer over the original input bytes. It emits tokens with:

- a surface string;
- a byte span into the original input;
- a token kind.

Normalization is used only for lookup-key derivation. It does not rewrite the stored token, the original input, or any byte span.

## Token classes

The tokenizer emits the following token kinds:

- `KindWord`: words and mixed alphanumeric runs containing at least one letter;
- `KindWhitespace`: one or more Unicode whitespace code points;
- `KindPunctuation`: punctuation code points, with specific handling for apostrophes and hyphen runs;
- `KindNumber`: numeric tokens under the number grammar below;
- `KindSymbol`: Unicode symbol code points;
- `KindUnknown`: invalid UTF-8 bytes or bytes/code points not classified elsewhere.

## Scanner rule

The scanner proceeds left to right and consumes the longest deterministic token at each position.

### Alphanumeric runs

ASCII letters and digits form a contiguous alphanumeric run. After the run is identified:

- if it contains at least one letter, the whole run is `KindWord`;
- if it contains only digits, it is `KindNumber`, subject to the number grammar.

This means examples such as `MP3`, `H2O`, `3rd`, and `21st` are word tokens.

### Whitespace

Whitespace is grouped into maximal whitespace runs.

### Punctuation

Punctuation is emitted as punctuation tokens, except where a specific rule below says it remains inside a word or is grouped.

### Symbols

Unicode symbols are emitted as symbol tokens.

### Unknown bytes

Invalid UTF-8 bytes are emitted one byte at a time as `KindUnknown`. Processing then continues at the next byte.

## Byte span policy

All spans are:

- zero-based;
- half-open;
- byte offsets into the original input string;
- independent of lookup normalization.

`TokenResult.Token` stores the original surface form for the token. The lookup key is a separate derived value.

## Lookup-key normalization

Lookup-key normalization consists of:

- lowercasing the token surface for dictionary and fallback lookup;
- treating ASCII apostrophe `U+0027` and right single quotation mark `U+2019` equivalently where the tokenizer recognizes an internal apostrophe.

No NFC, NFD, NFKC, or NFKD normalization is applied to the original input before tokenization. If a future release adds Unicode normalization to lookup keys, `NormalizerVersion` must change.

## Apostrophes

An apostrophe remains inside a word token only when it is between two Unicode letters.

Examples:

- `don't` is one word token;
- `it's` is one word token;
- `'hello'` is punctuation, word, punctuation;
- a leading or trailing apostrophe is punctuation.

The original apostrophe bytes remain unchanged in the stored token and span.

## Hyphens and dashes

ASCII hyphen-minus `U+002D` splits compound words:

```text
well-known → word "well", punctuation "-", word "known"
```

Multiple consecutive ASCII hyphens are grouped into one punctuation token:

```text
well--known → word "well", punctuation "--", word "known"
```

Other dash-like Unicode characters are punctuation tokens and are not treated as compound splitters.

A hyphen before a digit may be consumed as a number sign. A hyphen before a letter is punctuation.

## Number grammar

A number token has:

```text
number := sign? digits ( sep digits )?
sign   := "+" | "-"
digits := [0-9]+
sep    := "." | ","
```

There may be at most one internal separator. A second separator terminates the number token.

A sign belongs to a number only if it is immediately followed by a digit.

## Invalid UTF-8

Invalid UTF-8 is not rejected. Each invalid byte becomes a single-byte unknown token. The byte is preserved in the input span and receives a zero mask in `FeatureStream`.

## Versioning

The tokenizer and normalizer are versioned independently:

- `TokenizerVersion`, currently `v0.1`;
- `NormalizerVersion`, currently `v0.1`.

These versions are reported in result metadata and golden records. Any change that affects token boundaries, token kinds, lookup keys, or span behavior requires a version update and golden review.

## Alternatives considered

### Normalize input before tokenization

Rejected. Unicode normalization can change byte length and would make original byte spans ambiguous.

### Split mixed alphanumerics into numbers and words

Rejected. A single maximal word rule is simpler, deterministic, and avoids special cases such as `MP3` versus `3rd`.

### One punctuation token per hyphen

Rejected for repeated ASCII hyphens. Grouping a run such as `--` as one punctuation token is deterministic and matches common dash-like usage.

## Consequences

- Token spans are stable and reproducible.
- Lookup normalization cannot corrupt byte alignment.
- Mixed alphanumeric tokens are handled consistently.
- Golden tests can pin exact token boundaries and token kinds.

