package g2p_test

import (
	"fmt"

	"github.com/NK8007/gofonix/pkg/g2p"
)

// sourceName renders a g2p.Source for stable example output (the integer enum
// is intentionally avoided because consumers should read sources by name).
func sourceName(s g2p.Source) string {
	switch s {
	case g2p.SourceDict:
		return "dict"
	case g2p.SourceRuleFallback:
		return "rule-fallback"
	case g2p.SourceUnknown:
		return "unknown"
	}
	return "?"
}

// Example_basicUsage shows the canonical ModeBatch flow: construct an Engine
// with default options, call Process on a short English phrase, and inspect
// the tokens and their pronunciation provenance. The byte-aligned
// FeatureStream always has exactly len(Input) entries (ADR-0007).
func Example_basicUsage() {
	eng, err := g2p.New(g2p.Options{Language: "en", Mode: g2p.ModeBatch})
	if err != nil {
		panic(err)
	}

	res, err := eng.Process("hello world")
	if err != nil {
		panic(err)
	}

	fmt.Println("language:", res.Language)
	fmt.Println("tokens:", len(res.Tokens))
	fmt.Println("feature-stream:", len(res.FeatureStream), "(== len input:", len(res.Input) == len(res.FeatureStream), ")")
	for _, tok := range res.Tokens {
		fmt.Printf("  %-7q kind=%d source=%s phonemes=%d\n",
			tok.Token, tok.Kind, sourceName(tok.Pronunciation.Source),
			len(tok.Pronunciation.Phonemes))
	}
	fmt.Println("dict-id:", res.Metadata.DictionaryID)
	fmt.Println("schema:", res.SchemaVersion)

	// Output:
	// language: en
	// tokens: 3
	// feature-stream: 11 (== len input: true )
	//   "hello" kind=0 source=dict phonemes=4
	//   " "     kind=1 source=unknown phonemes=0
	//   "world" kind=0 source=rule-fallback phonemes=5
	// dict-id: cmudict-mini-v0.1
	// schema: gofonix-result-v0.1
}

// Example_oovHandling shows how an out-of-vocabulary word (here the nonce word
// "zymph") flows through the deterministic English rule fallback
// (fallback-en-v0.2, ADR-0009): the mini-dict misses, the fallback resolves
// the token, and Source becomes SourceRuleFallback rather than SourceDict.
// The FallbackRulesVersion metadata field exposes the frozen rule-set ID
// programmatically.
func Example_oovHandling() {
	eng, err := g2p.New(g2p.Options{Language: "en", Mode: g2p.ModeBatch})
	if err != nil {
		panic(err)
	}

	res, err := eng.Process("zymph")
	if err != nil {
		panic(err)
	}

	tok := res.Tokens[0]
	fmt.Println("token:", tok.Token)
	fmt.Println("source:", sourceName(tok.Pronunciation.Source))
	fmt.Println("phoneme-count:", len(tok.Pronunciation.Phonemes))
	fmt.Println("alignment-count:", len(tok.Pronunciation.Alignment))

	ids := make([]int, 0, len(tok.Pronunciation.Phonemes))
	for _, p := range tok.Pronunciation.Phonemes {
		ids = append(ids, p.ID)
	}
	fmt.Println("phoneme-ids:", ids)

	fmt.Println("oov-policy:", res.Metadata.OOVPolicy)
	fmt.Println("fallback-rules:", res.Metadata.FallbackRulesVersion)

	// Output:
	// token: zymph
	// source: rule-fallback
	// phoneme-count: 4
	// alignment-count: 4
	// phoneme-ids: [38 37 17 14]
	// oov-policy: rule-fallback-then-unknown
	// fallback-rules: fallback-en-v0.2
}

// Example_causalMode shows that ModeCausal is a scaffold (ADR-0003): every
// token is reported with SourceUnknown and the byte-aligned FeatureStream
// stays all-zero, even for tokens that would resolve cleanly under
// ModeBatch. A real streaming Causal simulator is deferred to a future
// release.
func Example_causalMode() {
	eng, err := g2p.New(g2p.Options{Language: "en", Mode: g2p.ModeCausal})
	if err != nil {
		panic(err)
	}

	res, err := eng.Process("hello")
	if err != nil {
		panic(err)
	}

	allZero := true
	for _, fm := range res.FeatureStream {
		if !fm.IsZero() {
			allZero = false
			break
		}
	}

	fmt.Println("mode-is-causal:", res.Mode == g2p.ModeCausal)
	fmt.Println("tokens:", len(res.Tokens))
	for _, tok := range res.Tokens {
		fmt.Printf("  %q source=%s phonemes=%d\n",
			tok.Token, sourceName(tok.Pronunciation.Source),
			len(tok.Pronunciation.Phonemes))
	}
	fmt.Println("feature-stream-len:", len(res.FeatureStream))
	fmt.Println("feature-stream-all-zero:", allZero)

	// Output:
	// mode-is-causal: true
	// tokens: 1
	//   "hello" source=unknown phonemes=0
	// feature-stream-len: 5
	// feature-stream-all-zero: true
}
