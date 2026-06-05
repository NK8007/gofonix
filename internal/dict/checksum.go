package dict

import (
	"crypto/sha256"
	"fmt"
)

// EmbeddedChecksum returns the hex-encoded SHA-256 of the embedded mini fixture
// data (ADR-0004). It is recorded in DATA_LICENSES.md and exposed via
// Result.Metadata.DictionaryChecksum so experiment manifests can pin the exact
// dictionary bytes used. The value is stable for a fixed fixture file.
func EmbeddedChecksum() string {
	h := sha256.Sum256(embeddedDictData)
	return fmt.Sprintf("%x", h)
}
