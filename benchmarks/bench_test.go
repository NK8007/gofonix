// Package benchmarks_test holds Gofonix's regression benchmarks (Slice 5).
//
// These benchmarks measure throughput and allocation behavior of the public
// g2p pipeline (and the internal dictionary lookup) over fixed fixtures. They
// deliberately set NO hard performance thresholds and report NO bits-per-byte
// or compression ratio — they exist purely to catch regressions over time.
//
// Run them with:
//
//	go test ./benchmarks/... -bench=. -benchmem -run=^$ -count=5
package benchmarks_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gofonix/gofonix/internal/dict"
	"github.com/gofonix/gofonix/pkg/g2p"
)

// fixturePath resolves a fixture file relative to this package directory. The
// fixtures live at <repo>/testdata/fixtures; the benchmarks package is one
// level below the repo root.
func fixturePath(name string) string {
	return filepath.Join("..", "testdata", "fixtures", name)
}

// loadFixture reads a fixture file or fails the benchmark.
func loadFixture(tb testing.TB, name string) string {
	tb.Helper()
	b, err := os.ReadFile(fixturePath(name))
	if err != nil {
		tb.Fatalf("read fixture %s: %v", name, err)
	}
	return string(b)
}

// benchmarkProcess runs Process over a fixture in a given mode, reporting
// allocations and bytes/op (input bytes processed per iteration).
func benchmarkProcess(b *testing.B, fixture string, mode g2p.Mode) {
	b.Helper()
	input := loadFixture(b, fixture)
	e, err := g2p.New(g2p.Options{Language: "en", Mode: mode})
	if err != nil {
		b.Fatalf("New: %v", err)
	}
	b.ReportAllocs()
	b.SetBytes(int64(len(input)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := e.Process(input)
		if err != nil {
			b.Fatalf("Process: %v", err)
		}
		// Touch a field so the call is not optimized away.
		if len(res.FeatureStream) != len(input) {
			b.Fatalf("FeatureStream len = %d, want %d", len(res.FeatureStream), len(input))
		}
	}
}

// BenchmarkBatchASCII measures ModeBatch over the ~10 KB ASCII fixture.
func BenchmarkBatchASCII(b *testing.B) {
	benchmarkProcess(b, "bench_ascii.txt", g2p.ModeBatch)
}

// BenchmarkBatchUnicode measures ModeBatch over the ~2 KB Unicode fixture
// (curly quotes, em/en dashes, accented letters).
func BenchmarkBatchUnicode(b *testing.B) {
	benchmarkProcess(b, "bench_unicode.txt", g2p.ModeBatch)
}

// BenchmarkCausalProcess measures ModeCausal over the ASCII fixture. The causal
// scaffold performs no lookup/projection in v0.1, so this is a useful baseline
// against the batch path.
func BenchmarkCausalProcess(b *testing.B) {
	benchmarkProcess(b, "bench_ascii.txt", g2p.ModeCausal)
}

// BenchmarkDictLookup measures the internal dictionary lookup hot path over a
// fixed set of in- and out-of-vocabulary keys. The internal/dict package is
// reachable from within the module, so this benchmark can exercise it directly.
func BenchmarkDictLookup(b *testing.B) {
	d, err := dict.LoadEmbedded()
	if err != nil {
		b.Fatalf("LoadEmbedded: %v", err)
	}
	keys := []string{"cat", "dog", "the", "it", "hello", "well", "known", "qwertyx", "nasa", "a"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k := keys[i%len(keys)]
		if _, ok := d.Lookup(k); ok {
			// hit
		}
	}
}
