package benchmarks_test

import (
	"testing"

	"github.com/gofonix/gofonix/internal/lang/en/fallback"
	g2p "github.com/gofonix/gofonix/pkg/g2p"
)

// Synthetic fixtures for the fallback benchmarks. These are inline string
// constants (not files) so the fallback benchmarks never touch the filesystem.
//
// oovHeavyText is dominated by out-of-vocabulary, fallback-eligible nonsense
// words; dictHeavyText is short common dictionary words; mixedText interleaves
// dictionary words, OOV words, digits, and punctuation.
const oovHeavyText = "zymph blarg fnord qux wumbo trix snarf yorp glom bweep " +
	"crumft splonk vrex noffle quib thrax wend zork blip skronk " +
	"plonk flib grox yulp snorf blund quaff twerp slink borf " +
	"zibble wonk splurt frex glib yomp blorf quidd snax vrunk"

const dictHeavyText = "the cat sat on the mat and the dog ran to the man " +
	"she said that he had a plan and it was good for all " +
	"we can see the sun and feel the air on a long clear day " +
	"time will show that the path is right if you try hard"

const mixedText = "the zymph ran 42 times and the blarg said hello! " +
	"fnord is a word but qux42 is not valid for fallback. " +
	"she can see the snarf near the vrex on a clear day, " +
	"yet wumbo and trix remain unknown to the dictionary."

// BenchmarkFallbackNormalizePerToken measures the cost of normalizing a single
// fallback-eligible surface form. It deliberately sets no b.SetBytes (per-token
// cost, not throughput).
func BenchmarkFallbackNormalizePerToken(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = fallback.Normalize("zymph")
	}
}

// BenchmarkFallbackMatchPerToken measures the cost of the longest-match scan
// over a single already-normalized token.
func BenchmarkFallbackMatchPerToken(b *testing.B) {
	normalized, ok := fallback.Normalize("zymph")
	if !ok {
		b.Fatal("normalize failed")
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = fallback.Match(normalized)
	}
}

// BenchmarkFallbackPronouncePerToken measures the end-to-end fallback
// convenience entry point (Normalize + Match) for a single token.
func BenchmarkFallbackPronouncePerToken(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = fallback.Pronounce("zymph")
	}
}

// benchmarkProcessInline runs Process over an inline string fixture in a given
// mode, reporting allocations and input bytes/op.
func benchmarkProcessInline(b *testing.B, input string, mode g2p.Mode) {
	b.Helper()
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
		if len(res.FeatureStream) != len(input) {
			b.Fatalf("FeatureStream len = %d, want %d", len(res.FeatureStream), len(input))
		}
	}
}

// BenchmarkProcessOOVHeavy measures ModeBatch over OOV-heavy text.
func BenchmarkProcessOOVHeavy(b *testing.B) { benchmarkProcessInline(b, oovHeavyText, g2p.ModeBatch) }

// BenchmarkProcessDictHeavy measures ModeBatch over dictionary-heavy text.
func BenchmarkProcessDictHeavy(b *testing.B) { benchmarkProcessInline(b, dictHeavyText, g2p.ModeBatch) }

// BenchmarkProcessMixed measures ModeBatch over mixed dictionary/OOV text.
func BenchmarkProcessMixed(b *testing.B) { benchmarkProcessInline(b, mixedText, g2p.ModeBatch) }

// BenchmarkProcessCausal measures ModeCausal over OOV-heavy text (same input as
// BenchmarkProcessOOVHeavy for comparability).
func BenchmarkProcessCausal(b *testing.B) { benchmarkProcessInline(b, oovHeavyText, g2p.ModeCausal) }
