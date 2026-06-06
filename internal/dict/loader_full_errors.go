package dict

import "errors"

// FullID is the dictionary identifier emitted in ResultMetadata.DictionaryID
// when the full CMUdict v0.7b has been successfully loaded and verified
// (ADR-0010). The mini-dict identifier remains EmbeddedID.
const FullID = "cmudict-full-v0.7b"

// FullExpectedSHA256 is the frozen hex-encoded SHA-256 digest the loader
// compares against (ADR-0010). Changing it requires bumping FullID (e.g. to
// "cmudict-full-v0.7c") and amending ADR-0010.
//
// NOTE: The placeholder value below is the SHA-256 of the empty string
// (e3b0c4429...). It is intentionally a value no real CMUdict file can have,
// because no CMUdict v0.7b distribution is empty. This guarantees that any
// real file on disk will fail the checksum check and trigger graceful
// fallback to the mini-dict (ADR-0010) until the true digest of the upstream
// artifact is pinned. The constant must not be edited locally to "match a
// copy on disk": ADR-0010 forbids any checksum bypass.
const FullExpectedSHA256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

// Sentinel errors returned by LoadFull (ADR-0010). The G2P engine treats every
// one of them as a graceful-degradation signal: log exactly one structured
// warning at WARN level, fall back to the embedded mini-dict, and continue
// construction successfully — never panic, never abort.
//
// Each error corresponds to one of the closed reason codes enumerated in
// ADR-0010 (Fallback policy). The engine maps the error to its reason string
// via FullDictFallbackReason.
var (
	// ErrFullDictNotBuilt is returned by the stub LoadFull when the binary was
	// built without the `gofonix_full_dict` build tag. This is the normal
	// default-build experience; the engine treats it as the silent "build tag
	// absent" case and emits no warning (ADR-0010, Fallback Policy).
	ErrFullDictNotBuilt = errors.New("dict: full CMUdict loader not built (missing -tags gofonix_full_dict)")

	// ErrFullDictPathUnresolved is returned when neither GOFONIX_DICT_PATH nor
	// $HOME yields a usable path to the full dictionary file (ADR-0010,
	// path resolution).
	ErrFullDictPathUnresolved = errors.New("dict: full CMUdict path unresolved (no GOFONIX_DICT_PATH and no $HOME)")

	// ErrFullDictFileUnreadable is returned when the resolved path does not
	// exist, is not a regular file, or cannot be opened for reading (ADR-0010).
	ErrFullDictFileUnreadable = errors.New("dict: full CMUdict file unreadable")

	// ErrFullDictSizeOutOfBand is returned when the file size falls outside
	// the [1 MiB, 16 MiB] sanity band (ADR-0010). This is a short-circuit
	// guard, NOT a substitute for the SHA-256 check.
	ErrFullDictSizeOutOfBand = errors.New("dict: full CMUdict size out of band (expected 1 MiB ≤ size ≤ 16 MiB)")

	// ErrFullDictChecksumMismatch is returned when the file's SHA-256 digest
	// does not match FullExpectedSHA256 (ADR-0010). There is no bypass flag.
	ErrFullDictChecksumMismatch = errors.New("dict: full CMUdict checksum mismatch (file does not match frozen cmudict-full-v0.7b digest)")

	// ErrFullDictEmptyAfterParse is returned when the parser produced zero
	// usable entries from the file's bytes — i.e. the file is present and
	// passes the checksum (impossible by construction) or, in tests, the
	// parser receives a non-CMUdict payload (ADR-0010).
	ErrFullDictEmptyAfterParse = errors.New("dict: full CMUdict parsed to zero entries")
)

// FullDictFallbackReason maps a LoadFull error to the closed reason code
// surfaced in the WARN log line emitted by the engine when the full-dict load
// fails (ADR-0010, Fallback Policy). Any unrecognised error collapses to
// "unexpected error" so the log always carries one of the closed values.
func FullDictFallbackReason(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, ErrFullDictNotBuilt):
		return "build tag absent"
	case errors.Is(err, ErrFullDictPathUnresolved):
		return "no path resolved"
	case errors.Is(err, ErrFullDictFileUnreadable):
		return "file unreadable"
	case errors.Is(err, ErrFullDictSizeOutOfBand):
		return "size out of band"
	case errors.Is(err, ErrFullDictChecksumMismatch):
		return "checksum mismatch"
	case errors.Is(err, ErrFullDictEmptyAfterParse):
		return "empty after parse"
	default:
		return "unexpected error"
	}
}
