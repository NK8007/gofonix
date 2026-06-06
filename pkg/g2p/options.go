package g2p

// Mode selects the information-availability semantics of a Process call
// (ADR-0003). In v0.1 ModeCausal is scaffold-only by design and produces an
// all-zero FeatureStream (ADR-0001, ADR-0003).
type Mode int

const (
	// ModeBatch has the full input text available. Analysis/upper-bound only;
	// not a fair streaming result. This is the default.
	ModeBatch Mode = iota
	// ModeOracle has full word and surrounding context available, including
	// bytes after the current position. An upper bound only.
	ModeOracle
	// ModeCausal may use only the decoded prefix. The only mode whose numbers
	// are admissible as fair byte-by-byte compression results. Scaffold-only in
	// v0.1 (all-zero FeatureStream).
	ModeCausal
)

// Options configures a new Engine (ADR-0001).
type Options struct {
	// Language is the BCP-47-ish language tag. v0.1 supports only "en"; an empty
	// value defaults to "en". Any other value yields ErrUnsupportedLanguage.
	Language string
	// Mode selects the analysis mode. Defaults to ModeBatch (the zero value).
	Mode Mode
	// DictPath optionally overrides the full-CMUdict resolution path used by
	// New when the binary is built with `-tags gofonix_full_dict` (ADR-0010,
	// dictionary path resolution). It is the first resolution step: a non-empty
	// value is used directly, ahead of GOFONIX_DICT_PATH and the
	// $HOME/.gofonix/cmudict.dict default. The field is honoured only in tagged
	// builds; in default builds the full loader is a stub and DictPath is
	// silently ignored as the engine falls back to the embedded mini-dict
	// (ADR-0010, fallback policy: "build tag absent" path, no warning).
	DictPath string
}
