package g2p

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"

	"github.com/NK8007/gofonix/internal/dict"
	"github.com/NK8007/gofonix/internal/lang/en/arpabet"
	"github.com/NK8007/gofonix/internal/lang/en/fallback"
	"github.com/NK8007/gofonix/internal/tokenizer"
	"github.com/NK8007/gofonix/pkg/phoneme"
)

// Result-record and component version constants recorded in ResultMetadata
// (ADR-0001, ADR-0004, ADR-0005, ADR-0006, ADR-0008).
const (
	schemaVersion        = "gofonix-result-v0.1"
	gofonixVersion       = "v0.1.0"
	featureSchemaVersion = "v0.1-en"
	normalizerVersion    = "v0.1"
	tokenizerVersion     = "v0.1"
	// oovPolicy is the v0.2 OOV policy after ADR-0009 amended ADR-0005: English
	// KindWord dictionary misses are attempted through the deterministic rule
	// fallback before being declined to SourceUnknown.
	oovPolicy = "rule-fallback-then-unknown"
	// fallbackRulesVersion is the frozen English fallback rule-set identifier
	// surfaced in ResultMetadata when the fallback is active (ADR-0009). It must
	// match internal/lang/en/fallback.RulesVersion.
	fallbackRulesVersion = fallback.RulesVersion
)

// Engine is the immutable, concurrency-safe G2P engine (ADR-0001). All engine
// state is fixed after New returns; Process allocates its own working state and
// shares no mutable data across calls, so Process is safe for concurrent use.
type Engine struct {
	language string
	mode     Mode
	// dict is the CMUdict loaded once in New. By default it is the embedded
	// mini-dict (ADR-0004); under `-tags gofonix_full_dict` New attempts to
	// load the full CMUdict v0.7b from disk and falls back to the mini-dict on
	// any error (ADR-0010, Fallback Policy). Immutable after construction and
	// safe for concurrent Lookup.
	dict *dict.Dict
	// dictID identifies the active dictionary backend for ResultMetadata
	// (ADR-0010, Decision §5). It is one of exactly two values in v0.3:
	// dict.EmbeddedID ("cmudict-mini-v0.1") or dict.FullID ("cmudict-full-v0.7b").
	dictID string
	// dictChecksum is the SHA-256 hex digest of the active dictionary bytes,
	// surfaced in ResultMetadata.DictionaryChecksum (ADR-0004, ADR-0010).
	dictChecksum string
	// fullDictAvailable is the programmatic signal exposed via
	// ResultMetadata.FullDictAvailable (ADR-0010, Public API Impact §1):
	// true iff the full CMUdict was loaded and verified successfully for this
	// engine instance; false otherwise (default builds, any fallback path).
	fullDictAvailable bool
}

// New constructs an Engine from opts (ADR-0001).
//
//   - An empty Language defaults to "en".
//   - Any Language other than "en" yields ErrUnsupportedLanguage.
//   - A Mode outside {ModeBatch, ModeOracle, ModeCausal} yields ErrInvalidMode.
func New(opts Options) (*Engine, error) {
	lang := opts.Language
	if lang == "" {
		lang = "en"
	}
	if lang != "en" {
		return nil, ErrUnsupportedLanguage
	}
	switch opts.Mode {
	case ModeBatch, ModeOracle, ModeCausal:
		// valid
	default:
		return nil, ErrInvalidMode
	}
	d, dictID, dictChecksum, fullOK := loadDictionary(opts.DictPath)
	if d == nil {
		// The mini-dict load is the last-resort fallback and is bundled in the
		// binary via //go:embed (ADR-0004). It cannot reasonably fail in v0.3;
		// if it ever does, surface the error rather than continuing with a
		// half-constructed engine.
		return nil, ErrDictionaryLoad
	}
	return &Engine{
		language:          lang,
		mode:              opts.Mode,
		dict:              d,
		dictID:            dictID,
		dictChecksum:      dictChecksum,
		fullDictAvailable: fullOK,
	}, nil
}

// loadDictionary implements the v0.3 dictionary-selection policy (ADR-0010,
// Decision §2–7 and Fallback Policy). It first attempts dict.LoadFull(path);
// on any error it logs a single structured WARN line (except for the silent
// "build tag absent" default-build case) and loads the embedded mini-dict.
//
// Returns the active *dict.Dict, its DictionaryID, its SHA-256 checksum, and
// the FullDictAvailable flag. On a complete failure (mini-dict load also
// failing, which cannot happen in normal builds because the file is embedded)
// the returned *dict.Dict is nil and the caller surfaces ErrDictionaryLoad.
func loadDictionary(pathOverride string) (*dict.Dict, string, string, bool) {
	d, err := dict.LoadFull(pathOverride)
	if err == nil {
		return d, dict.FullID, fullDictChecksumOf(d), true
	}
	// Log every failure path EXCEPT the silent default-build case where the
	// binary was compiled without `-tags gofonix_full_dict` (ADR-0010,
	// Fallback Policy: "build tag absent" must not emit a warning, because
	// that is the normal default user experience and warning spam is harmful).
	if !errors.Is(err, dict.ErrFullDictNotBuilt) {
		reason := dict.FullDictFallbackReason(err)
		log.Printf("gofonix: full CMUdict not loaded, falling back to mini-dict (reason=%q, err=%v)", reason, err)
	}
	mini, miniErr := dict.LoadEmbedded()
	if miniErr != nil {
		return nil, "", "", false
	}
	return mini, dict.EmbeddedID, dict.EmbeddedChecksum(), false
}

// fullDictChecksumOf returns the SHA-256 hex digest pinned to FullID by
// ADR-0010 (Checksum Policy). The frozen expected digest is the authoritative
// identifier for the full-dict bytes; we surface it unchanged rather than
// recomputing over the in-memory parsed structure (which would be a different,
// less meaningful digest).
func fullDictChecksumOf(_ *dict.Dict) string {
	// The constant is hex-encoded already; verify it is well-formed defensively
	// so a future edit cannot ship a non-hex digest into ResultMetadata.
	if _, err := hex.DecodeString(dict.FullExpectedSHA256); err != nil {
		// This is unreachable in normal builds; emit a non-empty marker so a
		// regression is loud rather than silent.
		h := sha256.Sum256([]byte(dict.FullExpectedSHA256))
		return hex.EncodeToString(h[:])
	}
	return dict.FullExpectedSHA256
}

// Process tokenizes input and returns a Result (ADR-0001, ADR-0006, ADR-0007).
//
// Slice 4 behavior: for word tokens in ModeBatch/ModeOracle, the dictionary is
// consulted; a hit's RAW ARPAbet canonical pronunciation is stress-stripped and
// mapped to neutral phoneme.Phoneme values via internal/lang/en/arpabet, the
// token's Pronunciation carries Source SourceDict with those phonemes, AND a
// uniform byte-based per-phoneme Alignment is computed over the token's byte
// span (ADR-0007). Each phoneme's FeatureMask is then projected into the
// FeatureStream over its alignment span. Misses, mapping failures, non-word
// tokens, and ALL tokens in ModeCausal yield Source SourceUnknown with an empty
// Pronunciation and contribute only zero masks.
//
// The FeatureStream always has exactly len(input) entries. In ModeBatch and
// ModeOracle it is the annotation/upper-bound projection of dictionary
// pronunciations; in ModeCausal it stays ALL-ZERO by design — the causal
// scaffold performs no lookup, no projection, and no prefix simulation in v0.1
// (ADR-0003). Invalid UTF-8 is tokenized, never rejected, so Process does not
// return an error in v0.1.
func (e *Engine) Process(input string) (Result, error) {
	toks := tokenizer.Tokenize(input)

	tokens := e.buildTokenResults(input, toks)

	// FeatureStream: exactly len(input) entries. Defaults to the zero value
	// FeatureMask{} == phoneme.Zero() for every byte. ModeBatch/ModeOracle then
	// project dictionary pronunciations over their alignment spans; ModeCausal
	// leaves the stream all-zero (the causal scaffold never projects; ADR-0003).
	fs := make([]phoneme.FeatureMask, len(input))
	if e.mode != ModeCausal {
		projectFeatureStream(fs, tokens)
	}

	return Result{
		Input:         input,
		Language:      e.language,
		Mode:          e.mode,
		Tokens:        tokens,
		FeatureStream: fs,
		SchemaVersion: schemaVersion,
		Metadata: ResultMetadata{
			GofonixVersion:       gofonixVersion,
			FeatureSchemaVersion: featureSchemaVersion,
			NormalizerVersion:    normalizerVersion,
			TokenizerVersion:     tokenizerVersion,
			OOVPolicy:            oovPolicy,
			FallbackRulesVersion: e.fallbackRulesVersion(),
			DictionaryID:         e.dictID,
			DictionaryChecksum:   e.dictChecksum,
			FullDictAvailable:    e.fullDictAvailable,
		},
	}, nil
}

// buildTokenResults adapts internal tokenizer tokens into public TokenResults
// and resolves each token's pronunciation (ADR-0004, ADR-0005, ADR-0008).
//
// Resolution policy (Slice 3):
//   - ModeCausal: every token is SourceUnknown with an empty Pronunciation
//     (the causal scaffold performs no dictionary lookup in v0.1; ADR-0003).
//   - ModeBatch/ModeOracle word tokens: look up the lowercased surface in the
//     dictionary. On a hit whose canonical RAW ARPAbet maps cleanly to neutral
//     phonemes, emit SourceDict with those phonemes. On a miss or any mapping
//     failure (an unknown ARPAbet symbol), fall back to SourceUnknown.
//   - Non-word tokens (whitespace, punctuation, number, symbol, unknown) are
//     always SourceUnknown.
//
// For a dictionary-hit word token, the Pronunciation also carries a uniform
// byte-based Alignment over the token's [Start, End) byte span, one ByteSpan
// per phoneme (ADR-0007). Variant is always 0 in v0.1 (ADR-0004).
func (e *Engine) buildTokenResults(input string, toks []tokenizer.Token) []TokenResult {
	out := make([]TokenResult, len(toks))
	for i, t := range toks {
		kind := mapKind(t.Kind)
		span := ByteSpan{Start: t.Start, End: t.End}
		pron := e.resolvePronunciation(kind, t.Surface, span)
		out[i] = TokenResult{
			Token:         t.Surface,
			Span:          span,
			Kind:          kind,
			Pronunciation: pron,
		}
	}
	return out
}

// fallbackRulesVersion reports the fallback rule-set identifier to record in
// ResultMetadata (ADR-0009). The English fallback is active only in
// ModeBatch/ModeOracle; ModeCausal stays scaffold-only and never applies the
// fallback, so it reports the empty string ("fallback disabled / inactive").
func (e *Engine) fallbackRulesVersion() string {
	if e.mode == ModeCausal {
		return ""
	}
	return fallbackRulesVersion
}

// emptyPronunciation is the canonical SourceUnknown pronunciation: no phonemes,
// no alignment, variant 0 (ADR-0005).
func emptyPronunciation() Pronunciation {
	return Pronunciation{
		Phonemes:  []phoneme.Phoneme{},
		Alignment: []ByteSpan{},
		Variant:   0,
		Source:    SourceUnknown,
	}
}

// resolvePronunciation resolves a single token's pronunciation per the Slice 4
// policy documented on buildTokenResults. span is the token's byte range, used
// to compute the per-phoneme uniform byte alignment on a dictionary hit.
func (e *Engine) resolvePronunciation(kind TokenKind, surface string, span ByteSpan) Pronunciation {
	// Causal scaffold: no lookup in v0.1 (ADR-0003). Only word tokens in
	// ModeBatch/ModeOracle are ever resolved against the dictionary.
	if e.mode == ModeCausal || kind != KindWord {
		return emptyPronunciation()
	}
	if entry, ok := e.dict.Lookup(surface); ok {
		if phonemes, ok := arpabet.MapSequence(entry.Canonical); ok {
			// Dictionary hit wins outright: it is resolved before the fallback is
			// ever consulted, so a canonical entry always yields SourceDict
			// (ADR-0009 Activation Rules). Uniform byte alignment over the token
			// span, one span per phoneme (ADR-0007). Purely byte-based.
			return pronunciationFrom(phonemes, span, SourceDict)
		}
		// Dictionary hit with unmappable ARPAbet: a true dictionary entry exists,
		// so this is NOT a dictionary miss. ADR-0009 fallback activation requires
		// a true miss; a corrupt/unmappable canonical entry stays SourceUnknown.
		return emptyPronunciation()
	}
	// Dictionary miss only: attempt the deterministic rule fallback (ADR-0009).
	// Pronounce normalizes and runs the longest-match scan, returning internal
	// ARPAbet symbols and an eligibility flag.
	syms, eligible := fallback.Pronounce(surface)
	if !eligible {
		// Fallback declined (empty/no-letters/digits/non-ASCII/unsupported):
		// unknown-only behaviour is retained (ADR-0009 Failure Cases).
		return emptyPronunciation()
	}
	// Map the internal ARPAbet symbols to neutral phonemes via the SAME bridge
	// dictionary hits use (ADR-0009 Output Semantics). ARPAbet never leaks out.
	phonemes, ok := arpabet.MapSequence(syms)
	if !ok {
		// The frozen table is closed to the 39-symbol inventory, so this cannot
		// happen for fallback-en-v0.2; decline defensively if it ever does.
		return emptyPronunciation()
	}
	return pronunciationFrom(phonemes, span, SourceRuleFallback)
}

// pronunciationFrom builds a resolved Pronunciation from a neutral phoneme
// sequence and the token's byte span. It computes the ADR-0007 uniform byte
// alignment (one span per phoneme) used identically for dictionary hits and
// rule-fallback resolutions; Variant is always 0 in v0.2 (ADR-0009).
func pronunciationFrom(phonemes []phoneme.Phoneme, span ByteSpan, src Source) Pronunciation {
	alignment := uniformByteAlignment(span.Start, span.End, len(phonemes))
	return Pronunciation{
		Phonemes:  phonemes,
		Alignment: alignment,
		Variant:   0,
		Source:    src,
	}
}

// boundaryPhonemeID is the universal boundary / non-speech phoneme ID
// (ADR-0008). Its FeatureMask (FeatBoundary, lo=4194304) is NOT emitted into
// the FeatureStream in v0.1: the stream must distinguish a non-speech byte
// (zero mask) from a boundary phoneme. The v0.1 dictionary path never produces
// ID 0, but the guard is explicit so projection can never leak a boundary mask.
const boundaryPhonemeID = 0

// projectFeatureStream writes the Batch/Oracle annotation FeatureStream in
// place (ADR-0007). fs has exactly len(input) entries, pre-initialized to the
// zero mask. For every token whose Pronunciation carries phonemes and a
// matching Alignment, each phoneme's v0.1-en FeatureMask is written over its
// alignment span [s, e):
//
//   - non-speech tokens, OOV words, unknown tokens, numbers, punctuation,
//     whitespace and symbols carry no phonemes, so they leave their bytes at
//     the zero mask;
//   - a zero-width span [k, k) writes nothing (the alignment entry still
//     exists on the Pronunciation, but it projects no bytes);
//   - the boundary phoneme (ID 0) is never emitted (boundary vs zero-mask
//     disambiguation).
//
// The uniform partition is non-overlapping within a token and token spans never
// overlap, so each fs byte is written at most once; the span bounds are clamped
// defensively against len(fs) so a malformed alignment can never panic.
func projectFeatureStream(fs []phoneme.FeatureMask, tokens []TokenResult) {
	for _, tr := range tokens {
		pr := tr.Pronunciation
		// Only resolved pronunciations with a 1:1 phoneme/alignment pairing
		// project. Anything else (empty/SourceUnknown) contributes zero masks.
		if len(pr.Alignment) != len(pr.Phonemes) {
			continue
		}
		for i, ph := range pr.Phonemes {
			if ph.ID == boundaryPhonemeID {
				// Boundary phoneme is not projected in v0.1 (ADR-0007/0008).
				continue
			}
			span := pr.Alignment[i]
			s, end := span.Start, span.End
			if s < 0 {
				s = 0
			}
			if end > len(fs) {
				end = len(fs)
			}
			if s >= end {
				// Zero-width (or invalid) span: nothing to write.
				continue
			}
			mask := arpabet.FeatureFor(ph.ID)
			for b := s; b < end; b++ {
				fs[b] = mask
			}
		}
	}
}

// mapKind translates an internal tokenizer.Kind to the public TokenKind. The
// two enums are defined in the same order, but the mapping is explicit so a
// future divergence is caught at compile time rather than silently miscoded.
func mapKind(k tokenizer.Kind) TokenKind {
	switch k {
	case tokenizer.KindWord:
		return KindWord
	case tokenizer.KindWhitespace:
		return KindWhitespace
	case tokenizer.KindPunctuation:
		return KindPunctuation
	case tokenizer.KindNumber:
		return KindNumber
	case tokenizer.KindSymbol:
		return KindSymbol
	default:
		return KindUnknown
	}
}
