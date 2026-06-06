# ADR-0001: Public API

## Status

Accepted.

## Context

Gofonix is a deterministic grapheme-to-phoneme and phonological feature extraction library. It is intended to support byte-level compression experiments while remaining usable as a small Go library and command-line tool.

The public API must serve two goals:

- expose stable, reproducible G2P results for experiments;
- keep the design English-only in implementation today, but multilingual-ready in architecture.

English ARPAbet and CMUdict are implementation details. Public types should use neutral phoneme and feature abstractions rather than English-specific encodings.

The API also needs to distinguish analysis modes. Batch and oracle-style processing may use full-input information and are suitable for annotation or upper-bound analysis. Causal processing must respect the no-look-ahead requirement needed for fair byte-by-byte predictive compression.

## Decision

Gofonix exposes two public packages:

- `pkg/phoneme`: language-neutral phoneme and phonological feature types;
- `pkg/g2p`: the G2P engine, options, modes, source classification, spans, and result types.

Language-specific implementation details live under `internal/`. This includes ARPAbet mapping, English fallback rules, tokenization details, dictionary loading, and embedded dictionary assets.

The public API is batch-call oriented. A caller provides an input string to `Process` and receives a complete `Result`. The result may be produced in different modes, but a separately callable streaming API is not part of the current public surface.

## Public concepts

### Modes

The public `Mode` enum distinguishes three information regimes:

- `ModeBatch`: full input text is available. This is the normal annotation mode.
- `ModeOracle`: full future information may be available. This is reserved for upper-bound analysis and must not be treated as a fair streaming result.
- `ModeCausal`: only causally available information may be used. In the current implementation this mode is scaffold-only and emits zero feature masks for all byte positions.

The causal contract is defined by ADR-0003. Current causal output must not be used for headline causal compression claims until a non-trivial prefix-simulation implementation is added.

### Byte spans

`ByteSpan` is the canonical position type:

- spans are zero-based and half-open: `[Start, End)`;
- offsets are byte offsets into `Result.Input`, not rune indexes;
- the invariant `0 <= Start <= End <= len(Result.Input)` must always hold;
- an empty span `[i, i)` is valid;
- normalization must not change span meaning.

All positions in tokenization, pronunciation alignment, and feature projection are byte-based.

### Sources

Pronunciation provenance is represented by a source enum:

- `SourceDict`: the pronunciation came from the dictionary;
- `SourceRuleFallback`: the pronunciation came from deterministic rule fallback;
- `SourceUnknown`: no pronunciation was produced.

This replaces any probabilistic confidence score in the public contract. The current OOV policy is `rule-fallback-then-unknown`: dictionary lookup is tried first, deterministic fallback is used for supported OOV words, and unsupported tokens remain unknown.

### Token kinds

Tokens are classified as:

- word;
- whitespace;
- punctuation;
- number;
- symbol;
- unknown.

The exact tokenizer and normalization rules are specified in ADR-0006.

### Result metadata

Every result carries reproducibility metadata. The metadata is part of the public result contract and must be populated consistently by both the library and CLI output.

```go
type ResultMetadata struct {
    GofonixVersion       string
    FeatureSchemaVersion string
    NormalizerVersion    string
    TokenizerVersion     string
    OOVPolicy            string
    FallbackRulesVersion string
    DictionaryID         string
    DictionaryChecksum   string
    FullDictAvailable    bool
}
```

The `ResultMetadata` field semantics are:

- `GofonixVersion`: the Gofonix release/module version, for example `v0.3.1-alpha`;
- `FeatureSchemaVersion`: the phonological feature schema version, for example `v0.1-en`;
- `NormalizerVersion` and `TokenizerVersion`: the normalizer and tokenizer versions (ADR-0006);
- `OOVPolicy` and `FallbackRulesVersion`: the active OOV handling policy and fallback rule version;
- `DictionaryID` and `DictionaryChecksum`: the dictionary identity and checksum actually used;
- `FullDictAvailable`: whether the optional full-dictionary backend was available for this run.

The result-record schema version is **not** part of `ResultMetadata`. It lives on the top-level `Result` as the `SchemaVersion` field (for example `gofonix-result-v0.1`); see the Go API sketch below.

`Result.SchemaVersion` and `ResultMetadata.GofonixVersion` are intentionally separate. A patch release may change `GofonixVersion`, other metadata, or implementation details without changing the result-record schema version.

### Thread safety

After construction, `Engine` is safe for concurrent calls to `Process`. Engine configuration and dictionary data are immutable after `New` returns. Each `Process` call uses its own working state.

### Invalid UTF-8

Invalid UTF-8 bytes are not a fatal input error. They are tokenized as single-byte unknown tokens, preserve their original byte spans, and occupy `FeatureStream` positions with the zero mask.

`ErrInvalidInput` is reserved for future explicit input policies. It is not used for ordinary invalid UTF-8 tokenization.

## Go API sketch

```go
package g2p

import "github.com/NK8007/gofonix/pkg/phoneme"

type Mode uint8

const (
    ModeBatch Mode = iota
    ModeOracle
    ModeCausal
)

type Source uint8

const (
    SourceDict Source = iota
    SourceRuleFallback
    SourceUnknown
)

type TokenKind uint8

const (
    KindWord TokenKind = iota
    KindWhitespace
    KindPunctuation
    KindNumber
    KindSymbol
    KindUnknown
)

type ByteSpan struct {
    Start int
    End   int
}

type Pronunciation struct {
    Phonemes  []phoneme.Phoneme
    Features  []phoneme.FeatureMask
    Alignment []ByteSpan
    Source    Source
    Variant   int
}

type TokenResult struct {
    Token         string
    Span          ByteSpan
    Kind          TokenKind
    Pronunciation Pronunciation
}

type Result struct {
    Input         string
    Language      string
    Mode          Mode
    Tokens        []TokenResult
    FeatureStream []phoneme.FeatureMask
    SchemaVersion string
    Metadata      ResultMetadata
}

type Options struct {
    Mode     Mode
    Language string
}

type Engine struct {
    // opaque
}

func New(opts Options) (*Engine, error)
func (e *Engine) Process(input string) (Result, error)
```

## Consequences

- The public API remains small and deterministic.
- English implementation details do not leak into exported identifiers.
- Results carry enough metadata to reproduce or interpret experiment artifacts.
- Batch annotation and causal prediction are kept semantically separate.
- Pre-1.0 releases may still change exported API details, but persisted results are protected by explicit schema and metadata versions.

