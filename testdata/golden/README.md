# Golden JSONL corpus

This directory holds the canonical **golden fixtures** used by the golden
comparison tests in [`pkg/g2p/golden_test.go`](../../pkg/g2p/golden_test.go).
Each file is a [JSON Lines](https://jsonlines.org/) file: one frozen snapshot of
`g2p.Engine.Process(input)` per line, serialized by the `internal/golden`
package. The full record format is specified in
[`docs/golden_jsonl_schema.md`](../../docs/golden_jsonl_schema.md).

## Files

| File             | Cases covered |
|------------------|---------------|
| `words.jsonl`    | Plain dictionary words in batch mode: `cat`, `a`, `the`, `dog`, `it`. Pins phoneme IDs and per-byte feature masks against ADR-0008 (e.g. `cat` → IDs 20/2/31, lo 1028/2146305/260). |
| `atypical.jsonl` | Atypical surface forms: trailing punctuation (`cat.`), ASCII apostrophe (`don't`, a `dict` hit), curly apostrophe U+2019 (`it’s`, a `unknown` non-word), wrapping quotes (`'hello'`), single and double hyphen runs (`well-known`, `well--known`), an acronym (`NASA`, resolved by the rule fallback to `source: rule_fallback`, IDs 23/2/29/2), and a multi-phoneme word (`hello`). |
| `oov.jsonl`      | Out-of-vocabulary words in batch mode: `qwertyx` and `NASA`. Both are resolved by the deterministic English rule fallback (ADR-0009), so each carries `source: rule_fallback` with **non-empty** phonemes/alignment and a **non-zero** feature stream (`NASA` → IDs 23/2/29/2). |
| `fallback.jsonl` | The rule-fallback corpus: OOV words the fallback resolves (`ship`, `thing`, `phone`, `xenon`, `qwerty` → `source: rule_fallback`), tokens the fallback declines (`MP3`, `H2O`, `3rd`, `21st`, `café`, `it’s` → `source: unknown`), a dictionary hit (`cat` → `dict`), a fallback word with trailing punctuation (`ship.`), and the same `qwerty` in **causal** mode (→ `unknown`, since causal performs no lookup). |
| `nonword.jsonl`  | Non-word runs: numbers (`123`, `3.14`), mixed tokens (`3rd`, `MP3`), whitespace (`" "`), and punctuation (`.`). |
| `causal.jsonl`   | The same kinds of input in **causal** mode. The scaffold performs no lookup/projection in v0.1 (ADR-0003): the feature stream is all-zero and every token is `source: unknown`. |

## How the comparison works

For each file the test:

1. Re-runs `Process(input)` for every `(input, mode)` case in the suite.
2. Serializes each live result with `internal/golden` and reads back the stored
   `.jsonl` line.
3. Requires an **exact** `reflect.DeepEqual` match, line by line.
4. Additionally re-checks the schema invariants (feature-stream length, byte
   spans, zero masks for non-speech and fallback-declined tokens, all-zero
   causal stream, boundary mask absence, contiguous alignment for resolved
   `dict`/`rule_fallback` words).

A normal `go test` run **never** rewrites these files and **fails** on any
mismatch.

## Regenerating (updating) goldens

Only regenerate intentionally — e.g. after a deliberate, reviewed change to the
engine, dictionary, or feature schema. Two equivalent ways:

```sh
# From the repo root:
go test ./pkg/g2p/ -run 'TestGolden$' -update

# or via the environment variable (handy in CI scripts):
GOFONIX_UPDATE_GOLDEN=1 go test ./pkg/g2p/ -run 'TestGolden$'
```

Then **review the diff** (`git diff testdata/golden/`) before committing. A
golden change is a contract change: confirm the new phoneme IDs, masks, spans,
and metadata (including the dictionary checksum) are what you intended.

> The fixtures intentionally contain **no** RAW ARPAbet — only neutral phoneme
> IDs and opaque `{lo,hi}` feature words (ADR-0002, ADR-0008).
