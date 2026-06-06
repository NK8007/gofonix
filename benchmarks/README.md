# Benchmarks

Regression benchmarks for the Gofonix G2P pipeline. They
measure throughput and allocation behavior over fixed fixtures and exist purely
to catch regressions over time.

These benchmarks deliberately:

- set **no hard performance thresholds** (they never `Fail` on a slow run);
- report **no bits-per-byte (bpb)** and **no compression ratio** — Gofonix v0.1
  ships no compressor (that is out of scope here);
- call `b.ReportAllocs()` so allocation regressions are visible.

## Fixtures

The benchmarks read from [`../testdata/fixtures`](../testdata/fixtures):

| Fixture              | Size    | Content |
|----------------------|---------|---------|
| `bench_ascii.txt`    | ~10 KB  | Original synthetic, neutral/technical English prose (ASCII only). |
| `bench_unicode.txt`  | ~2 KB   | Original synthetic prose with curly quotes (U+2018/U+2019/U+201C/U+201D), em-dash (U+2014), en-dash (U+2013), and accented Latin letters (`café`, `naïve`, `résumé`, `façade`). |

Both fixtures are original to this project and are not copied from any external
source. They are generated deterministically by
[`../scripts/gen_fixtures.py`](../scripts/gen_fixtures.py):

```sh
python3 scripts/gen_fixtures.py
```

## Benchmarks

| Benchmark              | What it measures |
|------------------------|------------------|
| `BenchmarkBatchASCII`    | `Process` in `ModeBatch` over the ASCII fixture. |
| `BenchmarkBatchUnicode`  | `Process` in `ModeBatch` over the Unicode fixture. |
| `BenchmarkCausalProcess` | `Process` in `ModeCausal` over the ASCII fixture (scaffold baseline: no lookup/projection). |
| `BenchmarkDictLookup`    | The internal `dict.Dict.Lookup` hot path over a mix of in- and out-of-vocabulary keys. |

## Fallback benchmarks

The fallback benchmarks live in [`bench_fallback_test.go`](./bench_fallback_test.go).
Unlike the file-backed fixtures above, they use **inline string constants**
(`oovHeavyText`, `dictHeavyText`, `mixedText`) so they never touch the
filesystem at runtime.

| Benchmark                            | What it measures |
|--------------------------------------|------------------|
| `BenchmarkFallbackNormalizePerToken` | `fallback.Normalize` over a single fallback-eligible token. |
| `BenchmarkFallbackMatchPerToken`     | `fallback.Match` (longest-match scan) over a single normalized token. |
| `BenchmarkFallbackPronouncePerToken` | `fallback.Pronounce` end-to-end (Normalize + Match) for a single token. |
| `BenchmarkProcessOOVHeavy`           | `Process` in `ModeBatch` over OOV-heavy inline text. |
| `BenchmarkProcessDictHeavy`          | `Process` in `ModeBatch` over dictionary-heavy inline text. |
| `BenchmarkProcessMixed`              | `Process` in `ModeBatch` over mixed dictionary/OOV inline text. |
| `BenchmarkProcessCausal`             | `Process` in `ModeCausal` over OOV-heavy inline text. |

Notes:

- The per-token benchmarks (`*PerToken`) deliberately set **no** `b.SetBytes()`,
  so they report **no `MB/s`** column. They measure absolute per-call cost and
  allocations, not throughput.
- `BenchmarkProcessCausal` uses `oovHeavyText` (the same input as
  `BenchmarkProcessOOVHeavy`) so the causal path is directly comparable against
  the OOV-heavy batch path.

To run only the fallback benchmarks:

```sh
go test ./benchmarks/... -bench=Fallback -benchmem -run=^$ -count=5
```

## Running

```sh
go test ./benchmarks/... -bench=. -benchmem -run=^$ -count=5
```

- `-bench=.` runs every benchmark.
- `-benchmem` prints `B/op` and `allocs/op` (also enabled in-code via
  `b.ReportAllocs()`).
- `-run=^$` skips ordinary tests (there are none in this package, but this keeps
  the run benchmark-only).
- `-count=5` repeats each benchmark five times so you can eyeball variance or
  feed the output to `benchstat`.

To compare two revisions:

```sh
go test ./benchmarks/... -bench=. -benchmem -run=^$ -count=10 > old.txt
# ... make a change ...
go test ./benchmarks/... -bench=. -benchmem -run=^$ -count=10 > new.txt
benchstat old.txt new.txt
```

The `MB/s` column comes from `b.SetBytes(len(input))` and reports input bytes
processed per second; it is a throughput figure, **not** a compression metric.
