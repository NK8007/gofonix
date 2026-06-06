package phoneme

// FeatureSchema identifies a versioned phonological feature schema (ADR-0002,
// ADR-0008). Masks produced under different schema versions are not comparable
// and must not be mixed in a single analysis.
//
// It carries only the version string; the concrete bit-position table is the
// v0.1-en assignment defined in ADR-0008 (see the Feat* constants below).
type FeatureSchema struct {
	// Version is the schema version string, e.g. "v0.1-en".
	Version string
}

// FeatureSchemaVersion is the version string of the v0.1 English feature schema
// (ADR-0008). Masks produced under different schema versions are not comparable.
const FeatureSchemaVersion = "v0.1-en"

// Feature bit positions for schema v0.1-en (ADR-0008). These are the 23 feature
// bits (0-22) defined by the v0.1 English schema; bits 23-127 are reserved.
const (
	FeatVoiced      = 0
	FeatNasal       = 1
	FeatStop        = 2
	FeatFricative   = 3
	FeatAffricate   = 4
	FeatApproximant = 5
	FeatLateral     = 6
	FeatLabial      = 7
	FeatAlveolar    = 8
	FeatPalatal     = 9
	FeatVelar       = 10
	FeatGlottal     = 11
	FeatHigh        = 12
	FeatMid         = 13
	FeatLow         = 14
	FeatFront       = 15
	FeatBack        = 16
	FeatRounded     = 17
	FeatTense       = 18
	FeatDiphthong   = 19
	FeatRhotic      = 20
	FeatSyllabic    = 21
	FeatBoundary    = 22
)
