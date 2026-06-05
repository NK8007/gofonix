package dict

import _ "embed"

// embeddedDictData is the committed mini CMUdict fixture (ADR-0004),
// version "cmudict-mini-v0.1". The full CMUdict is never committed.
//
//go:embed data/cmudict-mini-v0.1.dict
var embeddedDictData []byte

// EmbeddedID is the version identifier of the embedded fixture (ADR-0004). It is
// recorded as Result.Metadata.DictionaryID by the public engine.
const EmbeddedID = "cmudict-mini-v0.1"

// LoadEmbedded parses the embedded mini fixture and returns an immutable Dict
// (ADR-0004).
func LoadEmbedded() (*Dict, error) {
	return Load(embeddedDictData)
}
