# ADR-0010: Full CMUdict Loading Policy

## Status

Accepted. Refines ADR-0004 for optional full-dictionary loading.

## Context

Gofonix ships with a small embedded mini CMUdict. The mini dictionary keeps the module small, hermetic, and easy to test, but it has limited vocabulary coverage. When a word is missing from the active dictionary, Gofonix may fall back to the deterministic OOV rules from ADR-0009.

For larger English experiments, users may want full CMUdict coverage so that more words resolve as `SourceDict` rather than `SourceRuleFallback`.

The full CMUdict file should not be committed to the repository by default because it is larger, has separate data-provenance concerns, and would increase the size of every checkout, module fetch, and binary using unconditional embedding.

Gofonix therefore supports full CMUdict as an opt-in local artifact, not as a required repository asset.

## Decision

Gofonix supports an optional full CMUdict backend for English.

The default backend remains:

```text
cmudict-mini-v0.1
```

The optional full backend is identified as:

```text
cmudict-full-v0.7b
```

The full backend is available only when explicitly enabled and successfully verified. If anything fails, Gofonix falls back to the mini dictionary and continues normally.

## Build tags

The full dictionary loader is enabled by the Go build tag:

```text
gofonix_full_dict
```

Default builds without this tag use the mini dictionary only.

An advanced embedding build may additionally use:

```text
gofonix_full_dict_embed
```

The embed variant is for controlled distribution scenarios. It must not require committing the full dictionary file to the repository.

## Dictionary path resolution

When `gofonix_full_dict` is enabled and the non-embedded loader is used, the loader resolves the dictionary path in this order:

1. if `GOFONIX_DICT_PATH` is set and non-empty, use that path;
2. otherwise, if `$HOME` is set, use `$HOME/.gofonix/cmudict.dict`;
3. otherwise, no path is resolved and the loader falls back to the mini dictionary.

The loader reads local files only. It never downloads the dictionary.

## Verification policy

The full dictionary must pass all checks before it is used:

- the resolved path must exist;
- the path must point to a regular readable file;
- the file size must be within the accepted sanity range;
- the file bytes must match the compiled-in SHA-256 digest for `cmudict-full-v0.7b`;
- parsing must produce at least one usable entry.

Checksum verification is mandatory. There is no trust-on-first-use behavior and no checksum bypass flag.

## Current alpha behavior

Some alpha builds may include the full-dictionary loading path while keeping the real full-dictionary digest intentionally unpinned. In such builds, the loader is wired but checksum-gated: a real CMUdict file will not be accepted until the expected digest is pinned in code and documented.

In that state:

- `FullDictAvailable` remains `false`;
- `DictionaryID` remains `cmudict-mini-v0.1`;
- normal default behavior is unaffected;
- tagged full-dictionary tests still exercise graceful fallback behavior.

Pinning the real digest should be treated as a separate change because it changes whether full-dictionary output can actually be produced.

## Loading pipeline

When full dictionary loading is enabled, the loader:

1. resolves the dictionary path;
2. opens the file;
3. checks the size band;
4. computes SHA-256 over exact file bytes;
5. compares against the expected digest;
6. parses CMUdict entries;
7. indexes usable entries;
8. publishes the dictionary for immutable read-only use.

The loader runs during engine construction or dictionary initialization. It is not a hot-path operation, is not retried, and does not watch files for changes.

## Fallback policy

If the full dictionary cannot be used, Gofonix falls back to the mini dictionary.

Fallback must happen for:

- absent build tag;
- unresolved path;
- unreadable file;
- size outside the accepted range;
- checksum mismatch;
- empty or unusable parse result;
- unexpected load or parse error.

The public constructor must not panic or fail solely because the full dictionary is unavailable. Full dictionary loading is an optional enhancement, not a required dependency.

Default builds without `gofonix_full_dict` should not warn. That is the normal configuration.

## Metadata

Result metadata reports which dictionary backend is active.

For the mini dictionary:

```text
DictionaryID = "cmudict-mini-v0.1"
FullDictAvailable = false
```

For a successfully loaded full dictionary:

```text
DictionaryID = "cmudict-full-v0.7b"
FullDictAvailable = true
```

The dictionary checksum reported in metadata must correspond to the dictionary actually used.

Callers should use `DictionaryID`, `DictionaryChecksum`, and `FullDictAvailable` rather than parsing logs.

## Lookup semantics

The full dictionary uses the same lookup boundary as the mini dictionary.

A full-dictionary hit produces:

- `SourceDict`;
- neutral phonemes through the existing English ARPAbet bridge;
- `Variant = 0`;
- normal ADR-0007 alignment;
- normal ADR-0008 feature masks.

If a word is missing from the active dictionary, the ADR-0009 fallback policy still applies. Full dictionary support changes dictionary coverage, not fallback semantics.

## Prohibitions

The full dictionary loader must not:

- perform runtime network access;
- execute subprocesses;
- use CGO;
- auto-download during `go get`, `go install`, `go build`, `go test`, or first run;
- commit the full dictionary file to the repository;
- provide a checksum bypass flag;
- publish a partially loaded dictionary;
- mutate the dictionary after publication;
- hot-reload the dictionary from disk.

## Test requirements

The implementation should test:

- default build behavior with no full dictionary tag;
- path resolution with `GOFONIX_DICT_PATH`;
- path resolution with `$HOME/.gofonix/cmudict.dict`;
- fallback when no path is available;
- fallback on unreadable file;
- fallback on size-band failure;
- fallback on checksum mismatch;
- parser behavior for valid and invalid entries;
- metadata values for mini and full backends;
- dictionary-hit provenance under the full backend;
- unchanged ADR-0007 alignment and ADR-0008 feature masks;
- concurrent read safety after dictionary publication.

Tests that require the full dictionary should be gated behind the relevant build tag or use controlled fixtures.

## Alternatives considered

### Commit full CMUdict to the repository

Rejected. It increases repository and module size, complicates data provenance, and makes accidental local edits more likely.

### Runtime download

Rejected. Gofonix should not depend on network access for library construction, tests, or command execution.

### Trust-on-first-use checksum

Rejected. It would make the first local file silently define the trusted artifact, undermining reproducibility.

### Checksum bypass flag

Rejected. A bypass would eventually undermine the strict dictionary identity contract.

### Hot reload

Rejected. Dictionary data should be immutable for the lifetime of the engine.

## Consequences

- Default users keep a small, hermetic module with no external dictionary setup.
- Advanced users can opt into larger dictionary coverage.
- Full-dictionary availability is explicit in metadata.
- Misconfigured full-dictionary environments degrade safely to the mini dictionary.
- Reproducibility depends on pinning the full dictionary digest before full-dictionary output is treated as authoritative.

