package g2p

import "github.com/gofonix/gofonix/pkg/phoneme"

// Source records the provenance of a token's pronunciation (ADR-0001,
// ADR-0005, ADR-0009). All three values are emitted from v0.2 onward: a
// dictionary hit yields SourceDict, a dictionary miss that the English rule
// fallback resolves yields SourceRuleFallback, and anything the fallback
// declines (including every non-word token and every token in ModeCausal)
// yields SourceUnknown.
type Source int

const (
	// SourceDict means the pronunciation was resolved from the active CMUdict
	// (the unnumbered/base entry; ADR-0004). The dictionary backend is either
	// the embedded mini-dict or, under -tags gofonix_full_dict, the full
	// CMUdict v0.7b (ADR-0010); both surface as SourceDict.
	SourceDict Source = iota
	// SourceRuleFallback means the deterministic English rule fallback
	// (fallback-en-v0.2, ADR-0009) resolved an OOV KindWord token in
	// ModeBatch/ModeOracle. The phonemes flow through the same
	// internal/lang/en/arpabet bridge dictionary hits use; ARPAbet never
	// leaks into the public API.
	SourceRuleFallback
	// SourceUnknown means no pronunciation was produced: an OOV word the
	// fallback declined, any non-word token (whitespace, punctuation, number,
	// symbol, unknown byte), or any token in ModeCausal (scaffold-only;
	// ADR-0003).
	SourceUnknown
)

// TokenKind classifies a token (ADR-0001, ADR-0006).
type TokenKind int

const (
	// KindWord is an alphanumeric run containing at least one Unicode letter.
	KindWord TokenKind = iota
	// KindWhitespace is a maximal Unicode White_Space run.
	KindWhitespace
	// KindPunctuation is Unicode punctuation (grouped for ASCII hyphen runs).
	KindPunctuation
	// KindNumber is a digits-only run with optional sign and one separator.
	KindNumber
	// KindSymbol is a single Unicode symbol code point.
	KindSymbol
	// KindUnknown is a single byte that fits no other class, including invalid
	// UTF-8.
	KindUnknown
)

// ByteSpan is a zero-based, half-open [Start, End) range of byte offsets into
// Result.Input (ADR-0001, ADR-0006, ADR-0007). Invariant:
// 0 <= Start <= End <= len(Result.Input). An empty span [i,i) is valid.
// Normalization never changes spans.
type ByteSpan struct {
	// Start is the inclusive byte offset.
	Start int
	// End is the exclusive byte offset.
	End int
}

// Pronunciation holds the phoneme sequence for a token together with its
// per-phoneme byte alignment (ADR-0001, ADR-0007, ADR-0009). For a resolved
// word in ModeBatch/ModeOracle, Phonemes is populated, Alignment has one
// ByteSpan per phoneme via uniform byte distribution, and Variant is 0;
// Source is SourceDict on a dictionary hit and SourceRuleFallback when the
// English fallback (fallback-en-v0.2) supplied the phonemes. For non-word
// tokens, OOV words the fallback declined, and every token in ModeCausal it
// is empty with Source SourceUnknown.
type Pronunciation struct {
	// Phonemes is the phoneme sequence (empty when Source is SourceUnknown).
	Phonemes []phoneme.Phoneme
	// Alignment has one ByteSpan per phoneme via uniform byte distribution
	// (ADR-0007); len(Alignment) == len(Phonemes). Empty when SourceUnknown.
	Alignment []ByteSpan
	// Variant is the selected pronunciation variant; always 0 in v0.1
	// (ADR-0004).
	Variant int
	// Source records pronunciation provenance.
	Source Source
}

// TokenResult is one token plus its pronunciation (ADR-0001).
type TokenResult struct {
	// Token is the exact surface substring input[Span.Start:Span.End].
	Token string
	// Span is the original-byte range of the token.
	Span ByteSpan
	// Kind classifies the token.
	Kind TokenKind
	// Pronunciation carries phonemes/alignment.
	Pronunciation Pronunciation
}

// ResultMetadata exposes reproducibility-relevant versions so experiment
// harnesses cannot silently omit them (ADR-0001, Cross-issue E).
type ResultMetadata struct {
	// GofonixVersion is the gofonix release version, e.g. "v0.1.0".
	GofonixVersion string
	// FeatureSchemaVersion is the phonological feature schema version (ADR-0008).
	FeatureSchemaVersion string
	// NormalizerVersion is the normalizer version (ADR-0006).
	NormalizerVersion string
	// TokenizerVersion is the tokenizer version (ADR-0006).
	TokenizerVersion string
	// OOVPolicy is the out-of-vocabulary policy (ADR-0005, amended by ADR-0009).
	// It is "unknown-only" for builds without fallback (v0.1) and
	// "rule-fallback-then-unknown" when the English fallback is active (v0.2).
	OOVPolicy string
	// FallbackRulesVersion identifies the frozen deterministic OOV fallback rule
	// set (ADR-0009). It is the empty string "" when the fallback is disabled or
	// inactive (v0.1), and "fallback-en-v0.2" when the English fallback is active
	// (v0.2). ARPAbet symbols emitted by the rules never leak into the public API.
	FallbackRulesVersion string
	// DictionaryID identifies the dictionary used (ADR-0004, ADR-0010). In v0.3
	// it is one of exactly two values: "cmudict-mini-v0.1" (default builds and
	// any fallback path) or "cmudict-full-v0.7b" (successful, checksum-verified
	// load of the external full CMUdict under -tags gofonix_full_dict).
	DictionaryID string
	// DictionaryChecksum is the SHA-256 of the dictionary used (ADR-0004,
	// ADR-0010 Checksum Policy). For the embedded mini-dict it is computed
	// over the embedded bytes; for the full dict it is the frozen expected
	// digest pinned to FullID by ADR-0010.
	DictionaryChecksum string
	// FullDictAvailable is the programmatic signal for "did the full CMUdict
	// load successfully?" (ADR-0010, Public API Impact §1). It is true only
	// when the engine was constructed under -tags gofonix_full_dict AND the
	// on-disk file passed every check (path resolution, size band, SHA-256,
	// parse). It is false in every other case — including all default builds.
	// Callers must not rely on log output; this flag is the sole programmatic
	// signal of graceful degradation.
	FullDictAvailable bool
}

// Result is the full output of Process (ADR-0001). FeatureStream is byte-aligned
// with exactly len(Input) entries (ADR-0007). In ModeBatch/ModeOracle it is the
// annotation projection of dictionary pronunciations; in ModeCausal it is
// all-zero (the causal scaffold never projects in v0.1; ADR-0003).
type Result struct {
	// Input is the original input string.
	Input string
	// Language is the resolved language ("en" in v0.1).
	Language string
	// Mode is the analysis mode used.
	Mode Mode
	// Tokens is the deterministic token sequence.
	Tokens []TokenResult
	// FeatureStream has exactly len(Input) entries, one per input byte.
	FeatureStream []phoneme.FeatureMask
	// SchemaVersion is the result-record schema version.
	SchemaVersion string
	// Metadata carries reproducibility versions.
	Metadata ResultMetadata
}
