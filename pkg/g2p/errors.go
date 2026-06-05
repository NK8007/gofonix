package g2p

import "errors"

// Public sentinel errors (ADR-0001). ErrInvalidInput is reserved for future
// encoding violations and is effectively never returned by v0.1 Process:
// invalid UTF-8 is tokenized as KindUnknown rather than rejected (ADR-0006).
var (
	ErrUnsupportedLanguage = errors.New("gofonix: unsupported language")
	ErrInvalidMode         = errors.New("gofonix: invalid mode")
	ErrInvalidInput        = errors.New("gofonix: invalid input")
	// ErrDictionaryLoad is returned from New only when BOTH dictionary backends
	// failed to load: the optional full CMUdict declined (the common case —
	// graceful, handled internally) AND the embedded mini-dict also failed to
	// parse, which cannot happen in normal builds because the mini-dict ships
	// via //go:embed (ADR-0004, ADR-0010). This sentinel exists so a corrupt
	// release artifact surfaces a loud, actionable error rather than a panic.
	ErrDictionaryLoad = errors.New("gofonix: dictionary load failed")
)
