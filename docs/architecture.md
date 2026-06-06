# Gofonix Architecture

This document is a one-page map of the Gofonix module tree and the ADRs that
freeze its public contracts. For per-symbol behaviour read the godoc in
`pkg/g2p` and `pkg/phoneme`; for design rationale read the relevant ADR.

## Module Map

```
gofonix/
├── pkg/
│   ├── g2p/                       Public API surface (frozen by ADR-0001).
│   │   ├── Engine, Options, Mode  Constructor + analysis mode (batch / oracle
│   │   │                          / causal-scaffold; ADR-0003).
│   │   ├── Process                Batch entry point: full input → Result.
│   │   ├── Result, TokenResult,   Tokens, per-phoneme byte alignment
│   │   │   Pronunciation,         (ADR-0007), byte-aligned FeatureStream,
│   │   │   ResultMetadata         and reproducibility metadata.
│   │   └── example_test.go        Runnable examples: basic usage, OOV
│   │                              handling, causal-mode scaffold.
│   │
│   └── phoneme/                   Language-neutral phonological units.
│       ├── Phoneme (ID, IPA)      Neutral IDs 1–39 = v0.1-en inventory;
│       │                          ID 0 = boundary; 40+ reserved (ADR-0008).
│       ├── Feature / FeatureMask  Opaque struct{bits [2]uint64} with
│       │                          IsZero / Lo / Hi / Equal / Has.
│       └── Schema (v0.1-en)       Frozen feature set FeatVoiced=0 …
│                                  FeatBoundary=22 (ADR-0008).
│
├── internal/                      No public API. Subject to change.
│   ├── dict/                      CMUdict backends (ADR-0004, ADR-0010).
│   │   ├── data/                  Embedded mini-dict
│   │   │                          (cmudict-mini-v0.1.dict).
│   │   ├── embed.go, parser.go,   Mini-dict loader + entries + SHA-256.
│   │   │   checksum.go, entry.go,
│   │   │   dict.go
│   │   ├── loader_full.go         Full CMUdict v0.7b loader. Build tag:
│   │   │                          gofonix_full_dict. Validates size band,
│   │   │                          SHA-256, parse before accepting the file.
│   │   ├── loader_full_stub.go    Default-build stub (build tag
│   │   │                          !gofonix_full_dict).
│   │   └── loader_full_errors.go  Sentinel errors + FullDictFallbackReason
│   │                              mapper. Surfaced as
│   │                              ResultMetadata.FullDictAvailable.
│   │
│   ├── lang/en/                   English-specific resolution layer.
│   │   ├── arpabet/               39-symbol ARPAbet → neutral phoneme ID
│   │   │                          bridge. Strings never leak to the public
│   │   │                          API (ADR-0008).
│   │   └── fallback/              Deterministic rule-based OOV fallback
│   │                              (fallback-en-v0.2; ADR-0009). Activates
│   │                              on KindWord tokens missed by the
│   │                              dictionary; emits SourceRuleFallback.
│   │
│   ├── tokenizer/                 Deterministic token classifier
│   │                              (ADR-0006): KindWord, KindWhitespace,
│   │                              KindPunctuation, KindNumber, KindSymbol,
│   │                              KindUnknown.
│   └── golden/                    Golden-corpus harness; schema is frozen
│                                  in docs/golden_jsonl_schema.md.
│
├── cmd/
│   └── gofonix-cli/               Stdlib-only CLI entry point
│                                  (v0.3.0-alpha). Flags: --language,
│                                  --mode, --input, --dict-path, --output,
│                                  --version. Uses pkg/g2p only.
│
├── benchmarks/                    Reproducible Go benchmarks against the
│                                  public API. Not a binary; no runtime
│                                  dependencies beyond stdlib + pkg/g2p.
│
├── docs/
│   ├── adr/                       Architecture Decision Records.
│   ├── architecture.md            This document.
│   └── golden_jsonl_schema.md     Golden-corpus JSONL schema.
│
└── testdata/                      Fixture inputs shared across tests.
```

### Data Flow (one call to `Engine.Process`)

```
input string
     │
     ▼
internal/tokenizer  ──► tokens (KindWord / Whitespace / Punctuation / …)
     │
     ▼
internal/dict       ──► mini-dict lookup (or, under -tags gofonix_full_dict,
                        the validated full CMUdict v0.7b)
     │
     ▼ (miss + KindWord)
internal/lang/en/fallback (fallback-en-v0.2)
     │
     ▼
internal/lang/en/arpabet  ──► neutral phoneme IDs (1–39, ID 0 reserved)
     │
     ▼
pkg/g2p alignment (ADR-0007: uniform per-phoneme byte spans)
     │
     ▼
pkg/g2p feature projection (v0.1-en FeatureMask; ADR-0008)
     │
     ▼
g2p.Result { Tokens, FeatureStream (len == len(input)), Metadata }
```

The full dictionary is opt-in only; on any validation failure the engine
degrades silently to the embedded mini-dict and reports
`ResultMetadata.FullDictAvailable = false`.

## ADR Index

The ADRs are the authoritative source for every frozen contract. The full
set lives in [`docs/adr/`](adr/); the two most recent records are
summarised below.

### ADR-0009 — Deterministic OOV Fallback for English v0.2
[`docs/adr/ADR-0009-deterministic-oov-fallback.md`](adr/ADR-0009-deterministic-oov-fallback.md)

Freezes `fallback-en-v0.2`: a deterministic rule-based pronunciation layer
that resolves English `KindWord` tokens missed by the mini CMUdict. The
fallback supplies ARPAbet symbols that are mapped through
`internal/lang/en/arpabet` to neutral phoneme IDs and surface as
`SourceRuleFallback` (never `SourceDict`). Amends ADR-0005, switching the
English OOV policy from *unknown-only* to *rule-fallback-then-unknown*; all
other prior ADRs remain unchanged. Tokens the fallback declines, and every
non-word token, stay `SourceUnknown`.

### ADR-0010 — Full CMUdict Loading Policy for English v0.3
[`docs/adr/ADR-0010-full-cmudict-loading-policy.md`](adr/ADR-0010-full-cmudict-loading-policy.md)

Freezes the opt-in full-CMUdict-v0.7b backend. The
loader is gated behind the `gofonix_full_dict` build tag, resolves a path
(explicit `Options.DictPath` first, then `GOFONIX_DICT_PATH`, then the
`$HOME/.gofonix/cmudict.dict` default), validates a size band and an SHA-256
digest against a pinned constant, parses the file, and only then takes over
from the embedded mini-dict. Every failure mode (missing file, wrong size,
wrong checksum, parse error, or default build) results in silent, graceful
degradation to the embedded mini-dict with
`ResultMetadata.FullDictAvailable = false`. The full CMUdict is never
committed to the repository (license boundary + size). Refines ADR-0004 for
English `KindWord` tokens; ADR-0009 and all other prior ADRs are unaffected.
