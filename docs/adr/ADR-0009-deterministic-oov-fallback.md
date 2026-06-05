# ADR-0009: Deterministic OOV Fallback for English v0.2

**Status:** Accepted

**Date:** 2026 (Phase 2, Slice 1 — design freeze)

**Supersedes / amends:** ADR-0005 (OOV policy: from *unknown-only* to *rule-fallback-then-unknown* for English `KindWord` tokens). All other Phase 1 ADRs (ADR-0001, ADR-0003, ADR-0004, ADR-0006, ADR-0007, ADR-0008) are **unchanged** and remain authoritative.

---

## Context

Phase 1 (v0.1) shipped a deterministic, dictionary-only English G2P core:

- **Tokenization (ADR-0006):** `KindWord`, `KindWhitespace`, `KindPunctuation`, `KindNumber`, `KindSymbol`, `KindUnknown`.
- **CMUdict mini lookup (ADR-0004):** a dictionary hit on the unnumbered/base entry yields `SourceDict`, `Variant = 0`.
- **Internal ARPAbet → neutral phoneme mapping (`internal/lang/en/arpabet/`):** 39 symbols, neutral IDs 1–39. ARPAbet strings never leak into the public API (Principle 3).
- **FeatureMask schema (ADR-0008):** `v0.1-en`, features `FeatVoiced=0 … FeatBoundary=22`, opaque `struct{ bits [2]uint64 }`.
- **Uniform byte alignment (ADR-0007):** `N` phonemes → `N` `ByteSpan`s, half-open `[Start, End)`, distributed uniformly over the token's byte span.
- **Batch/Oracle FeatureStream projection (ADR-0007):** phoneme → bytes, non-speech bytes get a zero mask; `len(FeatureStream) == len(Input)`.
- **`ModeCausal` scaffold-only (ADR-0003):** all-zero FeatureStream; never projects in v0.1/v0.2.
- **OOV policy v0.1 (ADR-0005):** *unknown-only* — a word outside the dictionary → `SourceUnknown`, empty phonemes, zero mask.
- **Source enum (ADR-0001):** `SourceDict=0`, `SourceRuleFallback=1` (reserved in v0.1), `SourceUnknown=2`.
- **Golden JSONL + conformance tests + benchmarks (Slice 5).**

The v0.1 behaviour means any English word not in the bundled mini-dictionary produces no pronunciation at all. For v0.2 we want a **deterministic, rule-based fallback** so that common out-of-vocabulary (OOV) English words receive a *plausible, reproducible* pronunciation instead of silence — without compromising determinism, build hermeticity, or the existing public contracts. The reserved enum value `SourceRuleFallback` was put in place in v0.1 precisely for this purpose and is now activated.

This ADR is a **design freeze** for Phase 2 / Slice 1. It specifies behaviour and the frozen `fallback-en-v0.2` rule table; no production Go code is defined here.

---

## Decision

Introduce a **deterministic rule-based grapheme-to-phoneme fallback for English** in v0.2, governed by the following invariants:

1. **English only.** No other language module is added. The Polish module is explicitly out of scope.
2. **Deterministic, rule-based only.** No neural/statistical G2P, no probabilistic output, no runtime network calls, no subprocess, no CGO. Same input → identical output, always.
3. **Fallback fires only for `KindWord` tokens** that miss the dictionary and satisfy the activation predicate: ≥1 ASCII letter `[a-z]`, zero decimal digits, every character is `[a-z]` or U+0027 (see Activation Rules).
4. **Output reuses the existing internal ARPAbet bridge.** Rules emit ARPAbet-like symbols drawn from the existing 39-symbol `v0.1-en` inventory; those go through `internal/lang/en/arpabet/` exactly as dictionary phonemes do. ARPAbet stays internal.
5. **Alignment and FeatureStream are unchanged.** Fallback phonemes use the ADR-0007 uniform byte alignment, and Batch/Oracle projection behaves identically to dictionary hits. `ModeCausal` remains scaffold-only and is never mixed with fallback in v0.2.
6. **Provenance:** dictionary hit → `SourceDict`; fallback-resolved word → `SourceRuleFallback`; anything the fallback declines → `SourceUnknown`. `Variant = 0` for all fallback output in v0.2.
7. **Reproducibility:** a versioned rule set `fallback-en-v0.2`, surfaced via a new `ResultMetadata.FallbackRulesVersion` field, with golden tests.

This amends ADR-0005 for English: the OOV policy for English `KindWord` tokens becomes **rule-fallback-then-unknown**. The `unknown-only` behaviour is retained for every token the fallback declines.

---

## Fallback Activation Rules

The fallback is attempted **only** when *all* of the following hold:

- The token's `Kind` is `KindWord`.
- The mode is `ModeBatch` or `ModeOracle` (never `ModeCausal`).
- The dictionary lookup found **no** unnumbered/base CMUdict entry for the (lowercased) surface form.

A token is eligible for fallback if and only if:
1. At least one ASCII letter `[a-z]` is present (after lowercasing).
2. Zero decimal digits `[0-9]` are present.
3. Every character is either an ASCII letter `[a-z]` or an ASCII apostrophe U+0027 (`'`).

Tokens that fail any condition → `SourceUnknown`. Examples: `MP3`, `H2O`, `3rd` → `SourceUnknown`.

The fallback is **never** attempted for:

- `KindPunctuation`, `KindWhitespace`, `KindNumber`, `KindSymbol`, `KindUnknown` (including invalid UTF-8 bytes).
- Words that already have a canonical dictionary hit (those stay `SourceDict`).
- **Mixed alphanumeric tokens** that contain digits intermixed with letters — e.g. `MP3`, `H2O`, `3rd`, `21st`. These remain `SourceUnknown` in v0.2. (Whether the tokenizer keeps such strings as a single `KindWord` or splits them is governed by ADR-0006 and is *not* redesigned here; regardless, any token still carrying interior digits when it reaches the fallback is declined.)

---

## Rule Representation

Fallback rules are a **static, ordered grapheme → ARPAbet-symbol-sequence table**. Conceptually:

```
grapheme (1+ letters)  →  [ ARPAbet symbol, ... ]   (each symbol must exist in the v0.1-en 39-symbol inventory)
```

Properties of the table:

- **Closed inventory.** Every right-hand-side ARPAbet symbol must be one of the 39 `v0.1-en` symbols (`AA AE AH AO AW AY B CH D DH EH ER EY F G HH IH IY JH K L M N NG OW OY P R S SH T TH UH UW V W Y Z ZH`). The fallback can never emit a symbol the ARPAbet bridge does not know.
- **Stress-free.** Output symbols carry no stress markers; this matches the existing stress-stripped bridge. (Stress prediction is out of scope.)
- **Longest-match capable.** The table contains multi-letter graphemes (digraphs, common vowel patterns) as well as single-letter fallbacks, so the matcher can always make progress.
- **Versioned and frozen.** The full table is identified by `fallback-en-v0.2`. It is a data table, not behaviour; changing it requires bumping the version string.

> This section defines the **normative, frozen** `fallback-en-v0.2` rule set. The table below is the complete, authoritative rule table for this version.

### Fallback rule table

**Digraphs (multi-letter consonant clusters):**

| grapheme | ARPAbet |
|----------|---------|
| `sh`     | `SH`    |
| `ch`     | `CH`    |
| `th`     | `DH`    |
| `ph`     | `F`     |
| `wh`     | `W`     |
| `ng`     | `NG`    |
| `ck`     | `K`     |

**Common vowel patterns:**

| grapheme | ARPAbet |
|----------|---------|
| `ee`     | `IY`    |
| `ea`     | `IY`    |
| `oo`     | `UW`    |
| `ai`     | `EY`    |
| `ay`     | `EY`    |
| `ou`     | `AW`    |
| `ow`     | `OW`    |
| `oi`     | `OY`    |
| `oy`     | `OY`    |

**Single consonants:**

| grapheme | ARPAbet | | grapheme | ARPAbet | | grapheme | ARPAbet |
|----------|---------|-|----------|---------|-|----------|---------|
| `b`      | `B`     | | `c`      | `K`     | | `d`      | `D`     |
| `f`      | `F`     | | `g`      | `G`     | | `h`      | `HH`    |
| `j`      | `JH`    | | `k`      | `K`     | | `l`      | `L`     |
| `m`      | `M`     | | `n`      | `N`     | | `p`      | `P`     |
| `q`      | `K`     | | `r`      | `R`     | | `s`      | `S`     |
| `t`      | `T`     | | `v`      | `V`     | | `w`      | `W`     |
| `x`      | `K S`   | | `y`      | `Y`     | | `z`      | `Z`     |

**Simple vowels (single-letter defaults):**

| grapheme | ARPAbet |
|----------|---------|
| `a`      | `AE`    |
| `e`      | `EH`    |
| `i`      | `IH`    |
| `o`      | `AO`    |
| `u`      | `AH`    |

> **Normative status:** This table constitutes the frozen `fallback-en-v0.2` rule set. Duplicate LHS keys are invalid. All RHS phoneme symbols must be members of the 39-symbol ARPAbet inventory defined in `internal/lang/en/arpabet/`. No additional rules may be added without incrementing the fallback rules version.

Notes on selected rules:
- `x → K S` shows that a single grapheme may emit **more than one** ARPAbet symbol.
- `th → DH` is the deliberate v0.2 simplification (no voiced/voiceless distinction; `TH` is reserved for a future refinement and is not required by the longest-match algorithm).
- `y` is treated as a consonant `Y`; vowel-`y` handling, if any, is a table-data decision and does not change the algorithm.

---

## Normalization

**Normalization pipeline (normative, applied in order):**
1. **Lowercase** — Unicode fold to ASCII lowercase.
2. **Digit check** — if any `[0-9]` character is present → `SourceUnknown`, stop.
3. **Allowed-char check** — if any character is not `[a-z]` or U+0027 (`'`) → `SourceUnknown`, stop.
4. **Apostrophe strip** — remove all U+0027 characters (e.g. `don't` → `dont`).
5. **Empty check** — if the result is empty → `SourceUnknown`, stop.

Curly apostrophes (U+2018, U+2019) and non-ASCII characters such as `é` (U+00E9) are not ASCII letters and will be caught at step 3 → `SourceUnknown`. Thus `café` → `SourceUnknown` in v0.2.

---

## Longest-Match Algorithm

The matcher is a **deterministic greedy longest-match left-to-right scan** over the normalized, lowercased letter sequence of the token:

1. Start at position `i = 0`.
2. Among all table graphemes that match the input starting at `i`, choose the one with the **greatest length**.
3. **Tie-breaking** (same length): pick the grapheme that is **first alphabetically** by its grapheme key. This makes the choice fully deterministic and independent of map/iteration order.
4. Append the chosen grapheme's ARPAbet symbol sequence to the output, advance `i` by the matched grapheme length, and repeat.
5. If **no** grapheme matches at `i` (e.g. an unexpected character survived normalization), the entire token is **declined** → `SourceUnknown`. (Because the table includes every single base letter, this only happens for non-letter characters, which the activation rules already exclude.)
6. When `i` reaches the end of the input, the accumulated ARPAbet sequence is the fallback pronunciation.

Properties:

- **Total and terminating** over letter-only input: single-letter rules guarantee progress.
- **Deterministic**: longest-match + alphabetical tie-break + closed table ⇒ one canonical output per input.
- **No backtracking, no lookahead beyond the table's longest grapheme**: O(n · L) where L is the longest grapheme length (a small constant).

Apostrophes inside a word (e.g. `don't`-like OOV forms) are stripped during normalization for matching purposes and do not produce phonemes; they do not affect alignment spans, which remain over the original token bytes per ADR-0007.

---

## Output Semantics

- The longest-match scan produces a sequence of **internal ARPAbet-like symbols** drawn from the 39-symbol `v0.1-en` inventory.
- That sequence is mapped to **neutral `phoneme.Phoneme` values** by the existing `internal/lang/en/arpabet/` bridge — the *same* path dictionary pronunciations use. No new mapping code or inventory is introduced.
- **ARPAbet remains strictly internal.** Callers receive only neutral phonemes via the public `pkg/g2p` API (Principle 3 preserved).
- The resulting `Pronunciation` for a fallback-resolved word has: `Phonemes` populated, `Alignment` with one `ByteSpan` per phoneme (ADR-0007 uniform distribution), `Source = SourceRuleFallback`, `Variant = 0`.
- **Alignment and FeatureStream:** identical mechanics to dictionary hits. In `ModeBatch`/`ModeOracle`, each fallback phoneme's `v0.1-en` FeatureMask is projected over its alignment span; non-speech bytes stay zero. `ModeCausal` stays all-zero (scaffold-only); fallback is never applied in `ModeCausal`.

---

## Failure Cases

The fallback **declines** and the token yields `SourceUnknown` (empty `Phonemes`, empty `Alignment`, zero mask over its bytes) when:

- The token is **empty** after normalization.
- The token has **no letters** after normalization.
- The token is **mixed alphanumeric** — contains decimal digits together with letters (e.g. `MP3`, `H2O`, `3rd`, `21st`).
- The token contains **unsupported symbols** — characters outside ASCII letters + apostrophe (e.g. non-ASCII letters, diacritics, emoji) that survive normalization and have no matching grapheme.

Declining is the *safe* default: v0.2 never invents a pronunciation it cannot justify from the frozen letter-only rule table.

---

## Metadata and Reproducibility

- **Rule-set version:** `fallback-en-v0.2`. This identifies the exact frozen grapheme→ARPAbet table and the matcher semantics.
- **New metadata field:** `ResultMetadata.FallbackRulesVersion string`.
  - In v0.1 (and any build with the fallback disabled): **empty string** `""`.
  - In v0.2 with the English fallback active: `"fallback-en-v0.2"`.
- **Golden tests:** the golden JSONL corpus gains dedicated fallback cases (see Test Strategy), covering each emitted `SourceRuleFallback` pronunciation byte-for-byte.
- **Strict determinism guarantees:** no randomness, no runtime I/O, no network, no subprocess, no CGO, no user-specific or locale-specific state. The rule table is compiled-in static data; the matcher is pure.

---

## Public API Impact

The public surface changes are intentionally **minimal and additive**:

1. **`SourceRuleFallback` begins to be emitted.** The enum value already exists (ADR-0001, reserved in v0.1); no new type or constant is added. Consumers that already switch over `Source` must be prepared for this previously-unused value.
2. **`ResultMetadata.FallbackRulesVersion string`** is added — empty in v0.1, `"fallback-en-v0.2"` in v0.2.
3. The amended OOV policy (ADR-0005) is reflected in `ResultMetadata.OOVPolicy`: it is `"unknown-only"` for builds without fallback (v0.1) and `"rule-fallback-then-unknown"` when the English fallback is active (v0.2). These are the two closed, permitted values.

No existing field changes type or meaning. No public ARPAbet exposure. `Pronunciation`, `ByteSpan`, `TokenResult`, `Result`, and `FeatureStream` semantics are unchanged.

---

## Test Strategy

- **Unit tests for the longest-match algorithm:** greedy longest-match, alphabetical tie-breaking, progress/termination, multi-symbol emission (`x → K S`).
- **Fallback-table coverage tests:** at least one case exercising **every rule** in the frozen table (each digraph, vowel pattern, single consonant, simple vowel).
- **Provenance tests:**
  - OOV English word → `SourceRuleFallback` with the expected neutral phoneme sequence.
  - Dictionary hit → still `SourceDict` (no regression).
  - Non-word tokens (punctuation/whitespace/number/symbol/unknown) → still `SourceUnknown`.
  - Mixed alphanumeric (`MP3`, `H2O`, `3rd`, `21st`) → `SourceUnknown`.
- **Alignment / FeatureStream tests:** Batch/Oracle projection for a fallback word matches the ADR-0007 uniform-alignment contract, identical mechanics to a dictionary hit; `ModeCausal` stays all-zero.
- **Golden JSONL fallback cases:** new records pinning fallback pronunciations, alignment spans, sources, and `FallbackRulesVersion` byte-for-byte.
- **Benchmarks:** measure fallback cost vs the v0.1 dictionary-only baseline (per-token and end-to-end), confirming the fallback adds only bounded, deterministic overhead.

---

## Alternatives Considered

- **Neural / statistical G2P (e.g. seq2seq).** Rejected: non-deterministic across builds/hardware, heavyweight, conflicts with the no-CGO / no-runtime-I/O / hermetic-build constraints.
- **Probabilistic / weighted rule fallback.** Rejected: introduces non-determinism and ranking ambiguity; v0.2 wants one canonical output.
- **Shelling out to an external G2P tool (subprocess) or downloading a model at runtime.** Rejected: violates the no-subprocess, no-network constraints and breaks reproducibility.
- **Handling mixed alphanumeric (`MP3`, `3rd`) in v0.2.** Deferred: requires number-word expansion and ordinal/letter-name logic that is out of scope for this slice; kept `SourceUnknown`.
- **Voiced/voiceless `th` split (`TH` vs `DH`) and other context-sensitive refinements.** Deferred: the frozen `fallback-en-v0.2` table maps `th → DH` for simplicity; richer rules can land in a future `fallback-en-v0.3` without changing the algorithm.
- **Keeping v0.1 unknown-only and shipping no fallback.** Rejected as the goal of Phase 2 Slice 1 is precisely to provide a deterministic fallback; the reserved `SourceRuleFallback` enum exists for this.

---

## Consequences

**Positive:**
- Common English OOV words get a plausible, fully reproducible pronunciation instead of silence.
- Implementation is simple, auditable, and hermetic (static table + pure matcher).
- Zero new external dependencies; ARPAbet stays internal; public API change is additive.
- Determinism and golden-test discipline carry straight over from Phase 1.

**Negative / trade-offs:**
- Pronunciations are **approximate** — no morphology, no stress, no syllabification, no claim of linguistic correctness.
- Mixed alphanumeric and non-ASCII word material still produce no pronunciation (`SourceUnknown`).
- Consumers must now handle `SourceRuleFallback` in `Source` switches.
- The simplified `th → DH` mapping and similar choices are intentionally lossy for v0.2.

---

## Out of Scope

The following are **explicitly excluded** from ADR-0009 / v0.2:

- Neural G2P.
- Probabilistic fallback.
- Polish language module.
- Real causal-mode fallback (`ModeCausal` stays scaffold-only; no mixing).
- Any claim of full linguistic correctness.
- Morphology-aware English.
- Stress prediction beyond the simple no-stress policy.
- Syllabification.
- Dictionary learning / online adaptation.
- Runtime network downloads.
- Subprocess execution.
- CGO.

---

## Phase 2 Slice 1 Approval Checklist

Approve all items before proceeding to Slice 2.

- [ ] **Scope confirmed:** English-only, deterministic rule-based fallback; no neural/statistical/probabilistic, no network/subprocess/CGO.
- [ ] **Activation rules confirmed:** fires only for `KindWord`, dictionary-miss, operationally: ≥1 ASCII letter, zero digits, every char is `[a-z]` or U+0027; never for non-word kinds or canonical dict hits.
- [ ] **Mixed-alphanumeric policy confirmed:** `MP3`, `H2O`, `3rd`, `21st` → `SourceUnknown` in v0.2.
- [ ] **Source semantics confirmed:** dict → `SourceDict`; fallback → `SourceRuleFallback`; declined → `SourceUnknown`; `Variant = 0`.
- [ ] **Rule representation confirmed:** static, ordered grapheme → ARPAbet table; closed to the 39-symbol `v0.1-en` inventory; stress-free; frozen as `fallback-en-v0.2` (table is normative and part of this design freeze).
- [ ] **Longest-match algorithm confirmed:** greedy left-to-right, longest grapheme wins, alphabetical tie-break, single-letter rules guarantee progress, decline on no-match.
- [ ] **Output path confirmed:** rules emit internal ARPAbet → existing `internal/lang/en/arpabet/` bridge → neutral phonemes; ARPAbet never leaks to public API.
- [ ] **Alignment/FeatureStream confirmed:** ADR-0007 uniform byte alignment unchanged; Batch/Oracle projection identical to dict hits; `ModeCausal` all-zero, never mixed with fallback.
- [ ] **Failure cases confirmed:** empty / no-letters / mixed-alphanumeric / unsupported-symbol → `SourceUnknown`.
- [ ] **Metadata confirmed:** add `ResultMetadata.FallbackRulesVersion` (`""` in v0.1, `"fallback-en-v0.2"` in v0.2).
- [ ] **OOVPolicy string decided** for the fallback build: closed to exactly `"unknown-only"` (v0.1) or `"rule-fallback-then-unknown"` (v0.2).
- [ ] **Public API impact accepted:** `SourceRuleFallback` now emitted; additive metadata field only; no breaking changes.
- [ ] **Test strategy confirmed:** unit (matcher), full rule-table coverage, provenance, alignment/projection, golden JSONL fallback cases, benchmarks vs v0.1 baseline.
- [ ] **Out-of-scope list accepted** (neural, probabilistic, Polish, real causal fallback, correctness claims, morphology, stress beyond no-stress, syllabification, dictionary learning, network, subprocess, CGO).
- [ ] **ADR-0005 amendment accepted:** English OOV policy moves from *unknown-only* to *rule-fallback-then-unknown*; v0.1 ADR-0001/0003/0004/0006/0007/0008 unchanged.
