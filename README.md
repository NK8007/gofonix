# Gofonix

Deterministic grapheme-to-phoneme (G2P) engine for English with byte-aligned phonological feature streams, frozen by ADRs and reproducible across builds.

## Quick Start

```go
package main

import (
	"fmt"
	"github.com/NK8007/gofonix/pkg/g2p"
)

func main() {
	eng, _ := g2p.New(g2p.Options{Language: "en", Mode: g2p.ModeBatch})
	res, _ := eng.Process("hello world")
	fmt.Println(len(res.Tokens), len(res.FeatureStream), res.Metadata.DictionaryID)
}
```

## Installation

```sh
go get github.com/NK8007/gofonix@v0.3.0-alpha
```

Requires Go 1.22+. The default build is stdlib-only and has no runtime dependencies.

## Dictionary Setup

Gofonix selects its CMUdict backend at build time (ADR-0010):

- **Embedded (default).** Default builds ship the embedded mini CMUdict (`cmudict-mini-v0.1`). No setup is required; `ResultMetadata.FullDictAvailable` is always `false`.
- **Full CMUdict.** Build with `-tags gofonix_full_dict` and point the engine at an on-disk copy of CMUdict v0.7b via `Options.DictPath` (or let the loader resolve a default path). The file is validated by size band, SHA-256, and parse before being accepted; on any failure the engine degrades gracefully to the embedded mini-dict and sets `FullDictAvailable=false`. The dictionary file itself is never committed to this repository (`.gitignore`).

`ResultMetadata.DictionaryID` and `ResultMetadata.DictionaryChecksum` are the programmatic signals for which backend was actually used.

## CLI

```sh
go install github.com/NK8007/gofonix/cmd/gofonix-cli@v0.3.0-alpha
gofonix-cli --language en --mode batch --input "hello world" --output json
```

Flags: `--language`, `--mode {batch|causal}`, `--input`, `--dict-path`, `--output {json|text}`, `--version`. The CLI uses only the Go standard library.

## API Overview

The public API lives in [`pkg/g2p`](https://pkg.go.dev/github.com/NK8007/gofonix/pkg/g2p) and [`pkg/phoneme`](https://pkg.go.dev/github.com/NK8007/gofonix/pkg/phoneme):

- `g2p.Engine` — stateless analyzer constructed with `g2p.New(Options)`.
- `Engine.Process(input string) (Result, error)` — batch entry point; the full input is supplied in a single call.
- `g2p.Result` — tokens with per-phoneme byte alignment plus a `FeatureStream` with exactly `len(Input)` entries (ADR-0007) and a `Metadata` block carrying every reproducibility-relevant version.

Runnable examples for batch usage, OOV fallback handling, and the `ModeCausal` scaffold are available as `Example_basicUsage`, `Example_oovHandling`, and `Example_causalMode` in `pkg/g2p/example_test.go`.

## Architecture

See [`docs/architecture.md`](docs/architecture.md) for the module map and the ADR index.

## License

Licensed under the terms in [`LICENSE`](LICENSE) (file currently absent — see Compliance note in `CHANGELOG.md`). Third-party attributions, when added, will live in `NOTICE` / `THIRD_PARTY_NOTICES`.
