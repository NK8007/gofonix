# Gofonix (WIP)

Gofonix is an experimental, high-performance **Grapheme-to-Phoneme (G2P)** engine written in Go. The primary goal of this project is to provide deterministic transliteration of orthographic texts into the International Phonetic Alphabet (IPA) with an emphasis on extreme execution speed and zero I/O overhead.

> **Note:** This project is currently in the early Work In Progress (WIP) phase. The internal architecture and public API are subject to change.

## Project Objectives

While existing G2P libraries (such as Epitran for Python) are highly effective for analytical purposes, they often encounter performance bottlenecks when processing massive text corpora (gigabyte-scale). This is primarily due to the lack of native concurrency and reliance on external processes (e.g., Flite).

Gofonix addresses these limitations through:
- **Embedded Data:** Lexicons (e.g., CMU Dict) are embedded directly into the binary using `//go:embed`, eliminating I/O latency during runtime.
- **Native Concurrency:** Pipeline processing using Goroutines to fully utilize multi-core CPUs.
- **Feature Bitmasks:** Implementation of phonetic operations using fast bitwise masks instead of allocating heavy objects.
- **Portability:** The engine is designed to be compiled as a C-shared library (`-buildmode=c-shared`), enabling seamless integration with high-performance C/C++ systems.

## Project Structure

The repository architecture follows the standard [Go Project Layout](https://github.com/golang-standards/project-layout), maintaining a strict separation between internal mechanics and the public API.

* `cmd/gofonix-cli/` - The main command-line interface for testing, benchmarking, and profiling.
* `cmd/libgofonix/` - CGO wrappers exporting Gofonix as a shared library for C/C++ environments.
* `pkg/g2p/` - The core, public API for the transliteration engine.
* `pkg/phoneme/` - Representation of phonetic features.
* `internal/` - Encapsulated engine mechanics (including Trie data structures, embedded dictionaries, and contextual rules).
* `testdata/` - Test corpora used for performance benchmarking and validation.

## Development Status (Roadmap)

- [x] Initial architecture and package skeleton design.
- [ ] Implementation of the in-memory CMU dictionary lookup.
- [ ] Implementation of the pattern matching algorithm (Trie).
- [ ] Implementation of fallback rules for Out-Of-Vocabulary (OOV) words.
- [ ] Implementation of the worker pool with strict order determinism (Reorder buffer).
- [ ] Addition of CGO interfaces.

