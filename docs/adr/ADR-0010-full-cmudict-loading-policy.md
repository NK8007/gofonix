# ADR-0010: Full CMUdict Loading Policy for English v0.3

**Status:** Proposed

**Date:** 2026 (Phase 3, Slice 1 — design freeze)

**Supersedes / amends:** Refines ADR-0004 (CMUdict lookup policy) for English `KindWord` tokens by introducing an **opt-in, externally sourced full dictionary** alongside the bundled mini-dictionary. ADR-0009 (`fallback-en-v0.2`) and all other Phase 1/2 ADRs (ADR-0001, ADR-0003, ADR-0005, ADR-0006, ADR-0007, ADR-0008) are **unchanged** and remain authoritative.

---

## Context

Phase 1 (v0.1) shipped a deterministic, dictionary-only English G2P core built around a small, repository-bundled **mini-CMUdict** (a hand-curated subset of CMUdict v0.7b, on the order of a few hundred entries). Phase 2 (v0.2) layered a deterministic rule-based fallback on top (ADR-0009 / `fallback-en-v0.2`), so any English `KindWord` token outside the mini-dictionary that passes the activation predicate receives a plausible, reproducible pronunciation, and everything else stays `SourceUnknown`.

That topology is sufficient for tests and demos, but in real usage the mini-dictionary's coverage gap is large: most common English words still miss the dictionary and end up going through the rule-based fallback, which is intentionally approximate (ADR-0009, Consequences). For Phase 3 (v0.3) we want users who *opt in* to be able to ship Gofonix with the **full CMUdict v0.7b** (~130 000 entries, ~3.5 MB on disk) so that the vast majority of English words receive an authoritative dictionary pronunciation (`SourceDict`) rather than a fallback approximation (`SourceRuleFallback`).

However, the full CMUdict cannot be committed to the repository:

- **Size.** ~3.5 MB of dictionary text bloats the module, every `go get`, every container image, every CI clone, and every binary that uses `//go:embed` unconditionally. The mini-dict's whole point was to keep the module small and hermetic.
- **License / provenance.** CMUdict ships under the CMU/BSD-style license and must be redistributed with its own license text and attribution. Vending it inside the Gofonix module — which has its own license — creates an avoidable licensing-overlap problem that we prefer to push to the user/operator boundary.
- **Build hermeticity.** A vendored 3.5 MB asset is *the* artifact a future contributor would silently replace, regenerate, or "fix"; keeping the dictionary outside the repo with a fixed checksum makes the supply chain explicit.

At the same time, v0.3 must preserve every invariant from v0.1/v0.2: determinism, no runtime network, no subprocess, no CGO, the `internal/lang/en/arpabet/` bridge as the sole ARPAbet→neutral path, ADR-0007 uniform byte alignment, the `v0.1-en` FeatureMask schema, and the public API contract.

This ADR is a **design freeze** for Phase 3 / Slice 1. It specifies how the full CMUdict is loaded, identified, validated, and gracefully degraded. No production Go code is defined here.

---

## Decision

Introduce an **opt-in, externally sourced full CMUdict loading policy** for English in v0.3, governed by the following invariants:

1. **English only.** No other language module is added. Polish remains out of scope (consistent with ADR-0009).
2. **Opt-in via Go build tag.** The full dictionary is compiled in *only* when the build is invoked with `-tags gofonix_full_dict`. Default builds (no tag) keep shipping the mini-dictionary, byte-identical to v0.2. Default users incur zero size, zero license, and zero behaviour change.
3. **External, on-disk file — primary mechanism.** Under the `gofonix_full_dict` build tag, the dictionary text is read from a local filesystem path resolved as follows:
   - If the environment variable `GOFONIX_DICT_PATH` is set and non-empty, that exact path is used.
   - Otherwise, the default path is `~/.gofonix/cmudict.dict` (i.e. `$HOME/.gofonix/cmudict.dict`; on systems where `$HOME` is unset, this resolves to "no path" and triggers fallback per §Fallback Policy).
4. **Embedded variant — secondary, advanced mechanism.** A separately documented advanced opt-in (`gofonix_full_dict_embed`, a strict superset of `gofonix_full_dict` — implies it and is mutually-exclusive-by-precedence with the on-disk path) permits a conditional `//go:embed` of a dictionary file *that is checked in only as a `.gitignore`d local artifact*. The embed variant is offered for distribution scenarios where shipping a self-contained binary matters (CI images, single-binary releases) but is **not** the default and must never cause the large file to be committed to the repository.
5. **Stable dictionary identity.** When the full dictionary is successfully loaded and verified, its `DictionaryID` is the **fixed string** `"cmudict-full-v0.7b"`. The mini-dictionary's `DictionaryID` is unchanged (`"cmudict-mini-v0.1"`) and remains the default. The identifier is exposed in `ResultMetadata.DictionaryID` (already present since v0.1) so callers and golden tests can reason about which dictionary produced a given pronunciation.
6. **SHA-256 checksum is required.** The loader verifies the dictionary file against a frozen, compiled-in SHA-256 digest for `cmudict-full-v0.7b`. Any mismatch → loader declines → graceful fallback to mini (see §Fallback Policy). No "trust on first use", no "warn but use anyway".
7. **Graceful degradation, never panic.** If the file is missing, unreadable, malformed, checksum-mismatched, or the build tag is absent, the runtime **must** fall back to the mini-dictionary and log exactly one structured warning. The library never panics, never errors out of construction, never exits the process.
8. **No runtime network, no subprocess, no CGO, no auto-download.** The loader reads exactly one local file. Period.
9. **No change to lookup semantics, alignment, FeatureStream, or fallback algorithm.** The full dictionary slots into the same `dict.Lookup(word) (Pronunciation, ok)` boundary the mini-dict uses today (ADR-0004). ADR-0007 alignment, ADR-0008 FeatureMask `v0.1-en`, and ADR-0009 `fallback-en-v0.2` are byte-for-byte unchanged.
10. **Reproducibility.** The full dictionary is uniquely identified by `cmudict-full-v0.7b` + its frozen SHA-256. Two contributors with the same source file produce byte-identical golden output; a contributor with a different file is rejected at load time, not silently absorbed.

This **refines** ADR-0004 for English: the dictionary backend is now selectable at build time between `cmudict-mini-v0.1` (default) and `cmudict-full-v0.7b` (opt-in, externally sourced). The OOV policy from ADR-0005/ADR-0009 (`rule-fallback-then-unknown` in builds with the v0.2 fallback) is unchanged and continues to apply *after* whichever dictionary backend is active.

---

## Loading Pipeline

The loader's behaviour is normative. Applied **in this exact order** under `-tags gofonix_full_dict`:

1. **Resolve path.**
   - If `GOFONIX_DICT_PATH` is set and non-empty → use it verbatim.
   - Else if `$HOME` is set and non-empty → use `$HOME/.gofonix/cmudict.dict`.
   - Else → no path; skip to §Fallback Policy with reason `"no path resolved"`.
2. **Stat / open.**
   - If the file does not exist, is not a regular file, or cannot be opened → fallback with reason `"file unreadable"` and the resolved path.
3. **Size sanity bound.**
   - Reject files outside a closed size band: **1 MiB ≤ size ≤ 16 MiB**. (CMUdict v0.7b is ~3.5 MiB; the band exists purely to short-circuit obvious mistakes like an empty file, a misrouted log, or a multi-gigabyte mistake; it is **not** a security boundary and is **not** a substitute for the SHA-256 check.) Reject → fallback with reason `"size out of band"`.
4. **Streaming SHA-256.**
   - Compute SHA-256 of the file's exact bytes (no normalization, no line-ending translation). Compare against the frozen compiled-in digest for `cmudict-full-v0.7b`. Mismatch → fallback with reason `"checksum mismatch"`. The expected digest is treated as a constant of this ADR (see §Environment Setup).
5. **Parse.**
   - Parse the file according to the CMUdict v0.7b text format: one entry per line, `WORD  PH1 PH2 …` (ASCII, two-space separator in the canonical file), with `;;;` comment lines and variant suffixes `(2)`, `(3)`, … on duplicate headwords. Strip stress digits (`0`/`1`/`2`) at parse time, exactly as the mini-dict path already does (ADR-0004). Unknown ARPAbet symbols (outside the 39-symbol `v0.1-en` inventory) cause the affected **entry** to be skipped with a counted warning; they do not abort the load.
   - If the parser produces **zero usable entries**, treat the file as malformed → fallback with reason `"empty after parse"`.
6. **Index.**
   - Insert the parsed entries into the same in-memory dictionary structure used by the mini-dict (ADR-0004 lookup contract). On variant collision the canonical (unnumbered/base) entry wins for `Variant = 0` lookups, matching v0.1 semantics.
7. **Publish.**
   - Set `DictionaryID = "cmudict-full-v0.7b"` and make the dictionary live. Log exactly one structured info line `gofonix: loaded full CMUdict (entries=N, path=…, sha256=…)`. No further I/O for the lifetime of the process.

The loader is invoked **once**, at construction time (e.g. when the public `pkg/g2p` engine is first built). It is **never** invoked from a hot path, **never** retried, **never** watched for changes on disk.

---

## Fallback Policy

Whenever any step above declines, the loader **must**:

1. Continue construction successfully — never panic, never return a fatal error from the public constructor purely because the full dictionary could not be loaded.
2. Install the bundled mini-dictionary (`DictionaryID = "cmudict-mini-v0.1"`), exactly as in v0.1/v0.2 default builds.
3. Emit **exactly one** structured warning at WARN level, of the form:
   `gofonix: full CMUdict not loaded, falling back to mini-dict (reason=<reason>, path=<resolved or "">)`.
   The `reason` is one of the closed enumerated values: `"build tag absent"`, `"no path resolved"`, `"file unreadable"`, `"size out of band"`, `"checksum mismatch"`, `"empty after parse"`, `"unexpected error"`.
4. Set a single boolean flag on `ResultMetadata` (see §Public API Impact) so callers can detect graceful degradation without parsing logs.

Default builds (no `gofonix_full_dict` tag) take the same path with `reason = "build tag absent"` but emit **no warning** — that is the normal, expected default and warning spam is harmful. The metadata flag in §Public API Impact is the sole programmatic signal for the default case.

The v0.2 OOV pipeline (`rule-fallback-then-unknown`) runs unchanged on top of whichever dictionary won. Tokens missing the active dictionary continue to flow through `fallback-en-v0.2` (ADR-0009), and tokens the fallback declines stay `SourceUnknown`.

---

## Checksum Policy

- **Algorithm:** SHA-256, computed over the **exact file bytes** (no LF/CRLF normalization, no whitespace trimming, no comment stripping). This is the only safe digest semantics, because anything else allows two byte-different files to pass.
- **Expected digest:** compiled in as a string constant alongside the `gofonix_full_dict` tag. It is part of this ADR's design freeze; updating it requires bumping the dictionary identifier (e.g. `cmudict-full-v0.7c`) and amending this ADR. The concrete hex value of the digest is recorded in §Environment Setup so contributors can verify before they put the file in place.
- **On mismatch:** graceful fallback to mini per §Fallback Policy with `reason = "checksum mismatch"`. **Never** load the file anyway. **Never** prompt the user.
- **No "trust on first use".** Gofonix does not persist observed digests; the only trusted digest is the compiled-in one.

This is deliberately strict: a silent, slightly-modified CMUdict (e.g. with regional edits, or one stress digit altered) would silently change golden outputs and break reproducibility for everyone else on the project. Strict checksum enforcement makes that class of bug impossible to land.

---

## Environment Setup

The full CMUdict is an **external, contributor-provided** artifact. The following procedure is normative for any developer or operator who wants Gofonix to use it.

**One-time per machine:**

1. Obtain CMUdict v0.7b from the upstream source (Carnegie Mellon Speech Group's official distribution; the canonical filename is `cmudict.dict` or `cmudict-0.7b`). Place it at:
   `~/.gofonix/cmudict.dict`
   Alternatively, place it anywhere and set `GOFONIX_DICT_PATH=/absolute/path/to/cmudict.dict` in the build/run environment.
2. Verify the file's SHA-256 digest matches the value frozen for `cmudict-full-v0.7b` in the loader (e.g. `sha256sum ~/.gofonix/cmudict.dict` and compare to the digest constant carried by the build). If the digest differs, the loader will reject the file at runtime; do not proceed until the digest matches the upstream v0.7b file.
3. Ensure `~/.gofonix/cmudict.dict` (and any other on-disk copy of the dictionary) is covered by `.gitignore`. The full dictionary must never be committed to the Gofonix repository.

**Per build / per run:**

- Build with the tag: `go build -tags gofonix_full_dict ./...`
- Test with the tag: `go test -tags gofonix_full_dict ./...`
- Run with an override path (optional): `GOFONIX_DICT_PATH=/data/cmudict.dict ./your-binary`

**Embed variant (advanced, optional):**

- Build with both tags: `go build -tags 'gofonix_full_dict gofonix_full_dict_embed' ./...`
- The embed variant requires a `cmudict.dict` file placed at a fixed, documented in-tree path that is itself `.gitignore`d. The file is consumed via `//go:embed` at compile time. The SHA-256 check still runs against the embedded bytes; failure to match still triggers graceful fallback to mini.
- If both the embed and the on-disk file are usable in the same build, the embed wins (it is, by construction, hermetic). `GOFONIX_DICT_PATH` is ignored in pure-embed builds; this is intentional, because mixing the two is a determinism hazard.

**Default contributor experience (no tag):** unchanged. `go build ./...` and `go test ./...` work exactly as in v0.2, with the mini-dict, no warnings, no external files, no environment variables.

---

## Output Semantics

- A successful lookup against the **full** dictionary produces `SourceDict`, `Variant = 0`, and `ResultMetadata.DictionaryID = "cmudict-full-v0.7b"`, exactly as ADR-0004 specifies for any dictionary hit.
- A successful lookup against the **mini** dictionary (default builds, or any tagged build that fell back) is byte-identical to v0.2: `SourceDict`, `Variant = 0`, `DictionaryID = "cmudict-mini-v0.1"`.
- Dictionary misses continue to flow through the v0.2 fallback (ADR-0009), then to `SourceUnknown`. Provenance semantics are unchanged.
- Alignment (ADR-0007 uniform byte spans) and FeatureStream projection (ADR-0008 `v0.1-en`, Batch/Oracle only; `ModeCausal` all-zero) are unchanged.
- ARPAbet remains strictly internal: full-dictionary entries flow through `internal/lang/en/arpabet/` the same way mini-dict entries do (Principle 3 preserved).

---

## Metadata and Reproducibility

- **Dictionary identifier (existing field):** `ResultMetadata.DictionaryID`.
  - Default builds / fallback to mini: `"cmudict-mini-v0.1"`.
  - Successful full-dict load: `"cmudict-full-v0.7b"`.
  - These two values are the only ones permitted in v0.3.
- **New metadata field:** `ResultMetadata.FullDictAvailable bool`.
  - `true` when the full dictionary was loaded and verified successfully for this engine instance.
  - `false` otherwise (default builds, any fallback path).
  - This is the **sole** programmatic signal for "did the full dict load?". Callers must not rely on log output.
- **Fallback rules version (ADR-0009):** unchanged. Still `"fallback-en-v0.2"` in v0.2/v0.3 builds with the English fallback active; still `""` in builds without it.
- **OOV policy string (ADR-0005/0009):** unchanged. `"unknown-only"` or `"rule-fallback-then-unknown"`, exactly as in v0.2.
- **Strict determinism guarantees:** identical to v0.2 with one addition — once the loader has published a dictionary (full or mini), the dictionary is **immutable for the process lifetime**. No reloading, no watching, no concurrent mutation. Two runs of the same binary against the same on-disk file produce byte-identical golden output.

---

## Public API Impact

The public surface changes are intentionally **minimal and additive**:

1. **`ResultMetadata.FullDictAvailable bool`** is added — `false` in v0.1/v0.2 and in any v0.3 build where the full dict was not loaded; `true` only on successful, checksum-verified load.
2. **`ResultMetadata.DictionaryID`** (existing) gains a second permitted value `"cmudict-full-v0.7b"` alongside `"cmudict-mini-v0.1"`. No type change; consumers that already switch on it must add the new arm. This is the **only** way DictionaryID may legitimately differ between two builds in v0.3.
3. **No new exported types, no new constructors, no breaking changes** to `Pronunciation`, `ByteSpan`, `TokenResult`, `Result`, `FeatureStream`, `Source`, or `Mode`. `SourceRuleFallback` semantics from ADR-0009 are unchanged.
4. **No new exported environment variables beyond `GOFONIX_DICT_PATH`.** The two build tags (`gofonix_full_dict`, `gofonix_full_dict_embed`) are documented, not exported via the API.

---

## Failure Cases

The loader **declines** (and the mini-dict is used) when:

- The build was produced without `-tags gofonix_full_dict` (the normal default; no warning).
- `GOFONIX_DICT_PATH` is empty/unset and `$HOME` is unset → no path resolvable.
- The resolved path does not exist, is not a regular file, has no read permission, or the underlying read fails.
- The file size is outside the `[1 MiB, 16 MiB]` band.
- The SHA-256 digest does not match the frozen `cmudict-full-v0.7b` digest.
- The parser produces zero usable entries (file present but unrecognizable as CMUdict v0.7b text).
- Any unexpected I/O or parse error.

In every case the engine is constructed successfully, `FullDictAvailable = false`, `DictionaryID = "cmudict-mini-v0.1"`, and behaviour is byte-identical to v0.2.

---

## Categorical Prohibitions

The following are **forbidden** under this ADR and must be enforced both by the design and by reviewers:

- **No runtime network access.** No `net/http`, no `net/url` fetches, no `http.Get`, no DNS calls, no socket I/O, no proxy lookups, no implicit fetches via any third-party dependency. The loader's I/O budget is exactly one local file read.
- **No subprocess execution.** No `os/exec`, no shelling out to `curl`, `wget`, `python`, `sh`, or any external binary. The loader runs entirely in-process.
- **No CGO.** Consistent with every prior ADR. The loader is pure Go.
- **No auto-download on `go get`, `go build`, `go test`, or first run.** Module installation must not trigger any out-of-band fetch. The full dictionary is a **manual contributor action** described in §Environment Setup; it never happens implicitly.
- **No commit of the large dictionary file.** `cmudict.dict` (~3.5 MiB), any renamed copy of it, and any artifact derived from it that exceeds the mini-dict size must be `.gitignore`d. CI should treat the presence of such a file in a PR diff as a hard failure.
- **No "trust on first use" / no checksum bypass flag.** There is no `GOFONIX_DICT_SKIP_CHECKSUM`, no debug toggle, no "I know what I'm doing" override. Mismatched checksum → fallback to mini, always.
- **No silent partial loads.** If the loader cannot fully verify and index the dictionary, it does not publish a half-loaded structure — it falls back to mini.
- **No mutation of the loaded dictionary at runtime.** No add/remove/override APIs in v0.3. The dictionary is immutable for the process lifetime.
- **No locale-, time-, or user-dependent behaviour** beyond `$HOME` / `GOFONIX_DICT_PATH` resolution. Determinism (ADR-0009, Metadata and Reproducibility) is preserved.

---

## Test Strategy

Tests live behind the `gofonix_full_dict` build tag where they require the file to be present; everything else runs in the default build.

- **Default-build invariance:** with no tag, every existing v0.2 golden test and benchmark passes byte-identically. `FullDictAvailable == false`, `DictionaryID == "cmudict-mini-v0.1"`.
- **Path resolution:** `GOFONIX_DICT_PATH` honoured verbatim; falls back to `$HOME/.gofonix/cmudict.dict`; absent both → graceful fallback with the right reason.
- **Checksum:** a synthetic byte-altered fixture (a single byte flipped) triggers `"checksum mismatch"` and the mini-dict is used; a byte-identical good fixture loads successfully.
- **Size band:** sub-1 MiB and supra-16 MiB synthetic fixtures trigger `"size out of band"`.
- **Parser:** a CMUdict-shaped fixture containing a few entries with unknown ARPAbet symbols loads with those entries skipped (counted warning) but does not decline the file; an entirely non-CMUdict file declines with `"empty after parse"`.
- **Provenance under the full dict:** OOV-in-mini-but-in-full English words produce `SourceDict` + `"cmudict-full-v0.7b"` rather than `SourceRuleFallback` + `"fallback-en-v0.2"`; this is the primary user-visible v0.3 improvement.
- **Provenance under fallback:** a word missing from *both* dictionaries still flows to ADR-0009 fallback and, if declined, to `SourceUnknown`. ADR-0009 semantics unchanged.
- **Alignment / FeatureStream:** ADR-0007/0008 mechanics identical between mini-dict hits and full-dict hits for the same word (when both contain it).
- **Concurrency:** the dictionary is safe for concurrent reads after publication; no test or runtime path attempts to mutate it.
- **No-network / no-subprocess assertion:** a test that stubs out `os/exec` and `net` confirms no symbol from either is reachable in the loader's call graph (or, equivalently, a `go list -deps`-style audit on the loader package).
- **Benchmarks:** end-to-end pronunciation throughput under both backends, plus one-time load cost for the full dict (target: tens of milliseconds for the parse+hash, not seconds).

---

## Alternatives Considered

- **Vendoring CMUdict v0.7b into the repository.** Rejected: ~3.5 MiB module bloat for every consumer, license overlap with the Gofonix module license, and a tempting target for ad-hoc local edits that silently break determinism for everyone else.
- **Unconditional `//go:embed` of an in-tree dictionary.** Rejected for the default build for the same reasons; permitted only as an advanced opt-in (`gofonix_full_dict_embed`) where the file is `.gitignore`d and supplied by the contributor.
- **Runtime download (HTTP GET on first start, with cache).** Rejected categorically. Violates the no-runtime-network constraint that has held since ADR-0009, introduces non-determinism (mirror drift, network failures), and is a supply-chain liability. Listed in §Categorical Prohibitions for emphasis.
- **Lazy / streaming lookup against a memory-mapped file.** Rejected for v0.3: complicates the dictionary API surface, breaks the "dictionary is immutable in-memory data" mental model used by the existing `dict.Lookup` boundary, and saves a few MiB of RAM at the cost of every reviewer's sanity. Can be reconsidered if memory pressure ever becomes a real constraint.
- **Trust-on-first-use checksum policy.** Rejected: silently absorbs whatever file the first-running contributor happened to have, defeating the point of pinning to `cmudict-full-v0.7b`.
- **A `GOFONIX_DICT_SKIP_CHECKSUM` escape hatch.** Rejected: any such flag becomes the default in CI within a quarter and erodes the determinism contract. There is no checksum bypass.
- **Multiple full-dict variants (regional, custom).** Deferred. v0.3 freezes exactly one identifier (`cmudict-full-v0.7b`) with exactly one expected digest. A future ADR may introduce additional named variants, each with its own pinned digest.
- **Auto-fetch with explicit user consent (interactive prompt).** Rejected: Gofonix is a library, not an interactive tool; a library that prompts on import is the worst of all worlds.

---

## Consequences

**Positive:**
- Opt-in users get full CMUdict v0.7b coverage — the vast majority of English OOV words in v0.2 collapse onto authoritative `SourceDict` pronunciations.
- The default repository, default build, and default `go get` cost stay exactly where they were in v0.2: small, hermetic, no external dependencies, no license entanglements.
- Determinism is *strengthened*, not weakened: the SHA-256 pin makes "which dictionary is in play" a single, checkable string, and `FullDictAvailable`/`DictionaryID` make it programmatically inspectable.
- Graceful degradation means a misconfigured contributor never breaks the build; they just don't get the upgrade.

**Negative / trade-offs:**
- Opt-in users now have a manual setup step (download v0.7b, place it on disk, verify digest). This is intentional but unavoidably more friction than a vendored file.
- Two dictionary backends mean golden tests must distinguish "produced by mini" vs "produced by full" outputs and run under the right tag. The `DictionaryID` and the `FullDictAvailable` flag are the discriminators.
- The advanced embed variant exists and must be documented and tested even though most users will never touch it.
- A misconfigured CI (e.g. cache poisoning of `~/.gofonix/cmudict.dict`) will produce mini-dict output rather than failing loudly. This is the chosen graceful-degradation trade-off; the metadata flag is the audit signal CI is expected to check.

---

## Out of Scope

The following are **explicitly excluded** from ADR-0010 / v0.3:

- Runtime network downloads of any kind (HTTP, HTTPS, git, S3, anything).
- Subprocess execution to fetch, decompress, or transform the dictionary.
- CGO of any kind.
- Auto-download triggered by `go get`, `go install`, `go build`, `go test`, or first program start.
- Committing the full CMUdict (or any ≥1 MiB dictionary asset) to the Gofonix repository.
- A checksum bypass / "skip verification" flag in any form.
- Multiple full-dict variants beyond `cmudict-full-v0.7b` (e.g. regional or custom dictionaries).
- Hot reload / file-watching / live re-indexing of the dictionary.
- User-facing add/remove/override APIs on top of the loaded dictionary.
- Polish or any non-English language module.
- Neural / statistical G2P (already excluded by ADR-0009).
- Changes to ADR-0007 alignment, ADR-0008 FeatureMask, or ADR-0009 fallback rules.
- Changes to the public API beyond the single additive `FullDictAvailable` field and the second permitted `DictionaryID` value.
- Stress prediction, syllabification, morphology — still out of scope as in ADR-0009.

---

## Phase 3 Slice 1 Approval Checklist

Approve all items before proceeding to Slice 2.

- [ ] **Scope confirmed:** English-only, opt-in full CMUdict v0.7b loading; default builds unchanged from v0.2.
- [ ] **Build-tag gating confirmed:** `gofonix_full_dict` selects the on-disk loader; absent tag → mini-dict, no warning.
- [ ] **Path resolution confirmed:** `GOFONIX_DICT_PATH` overrides; default `$HOME/.gofonix/cmudict.dict`; neither resolvable → graceful fallback.
- [ ] **Dictionary identifier confirmed:** exactly `"cmudict-full-v0.7b"` on success; `"cmudict-mini-v0.1"` otherwise; these are the only two permitted values.
- [ ] **SHA-256 policy confirmed:** mandatory, frozen compiled-in digest; mismatch → graceful fallback; no bypass flag.
- [ ] **Size band confirmed:** `[1 MiB, 16 MiB]` sanity bound, separate from and *not* a substitute for the checksum check.
- [ ] **Graceful degradation confirmed:** never panic, never abort construction; exactly one warning at WARN level (except for the default-build "tag absent" case, which is silent); `FullDictAvailable` is the programmatic signal.
- [ ] **Embed variant confirmed as advanced opt-in:** `gofonix_full_dict_embed` requires a `.gitignore`d in-tree file; same SHA-256 enforcement; documented but not the default.
- [ ] **Public API impact accepted:** single new `ResultMetadata.FullDictAvailable bool`; one new permitted value for the existing `DictionaryID` field; no breaking changes.
- [ ] **Categorical prohibitions accepted:** no runtime network, no subprocess, no CGO, no auto-download, no commit of the big file, no checksum bypass, no partial loads, no runtime mutation.
- [ ] **Test strategy confirmed:** default-build invariance, path resolution, checksum (good + bad fixtures), size band, parser tolerance, provenance under both backends, ADR-0007/0008 invariance, concurrency, no-network/no-subprocess audit, benchmarks.
- [ ] **Environment setup procedure confirmed** and ready to be added to `CONTRIBUTING` / build documentation.
- [ ] **ADR-0004 refinement accepted:** dictionary backend becomes a build-time selection between `cmudict-mini-v0.1` (default) and `cmudict-full-v0.7b` (opt-in); ADR-0005, ADR-0007, ADR-0008, ADR-0009 unchanged.
- [ ] **Out-of-scope list accepted** (runtime network, subprocess, CGO, auto-download, committed big files, checksum bypass, extra variants, hot reload, mutation APIs, Polish, neural G2P, public-API changes beyond the single additive field).
