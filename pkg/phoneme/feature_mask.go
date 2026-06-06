package phoneme

import "encoding/json"

// FeatureMask is an opaque bitmask holding up to 128 phonological feature bits.
//
// It must NOT be a uint64 alias — it is a struct, so the mask stays opaque and
// future-extensible (ADR-0002). The v0.1-en schema (ADR-0008) uses only bits
// 0-22, all of which live in the low word; the high word is reserved for
// future schema versions.
type FeatureMask struct {
	bits [2]uint64
}

// MarshalJSON renders the mask as {"lo":<low 64 bits>,"hi":<high 64 bits>},
// matching the {lo,hi} feature-word form used by the golden JSONL schema
// (docs/golden_jsonl_schema.md) and internal/golden.Mask.
//
// Without this method encoding/json would emit an empty object {} for the
// mask because the underlying bits field is unexported, which silently dropped
// the feature values from any json.Marshal of a Result (e.g. the CLI's
// --output json). This is a serialization-only fix: it surfaces the existing
// Lo()/Hi() accessor values and does not change the mask, the feature schema,
// projection, or any bit semantics.
func (m FeatureMask) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Lo uint64 `json:"lo"`
		Hi uint64 `json:"hi"`
	}{Lo: m.bits[0], Hi: m.bits[1]})
}

// Zero returns the zero FeatureMask (no features set).
func Zero() FeatureMask { return FeatureMask{} }

// IsZero reports whether no features are set.
func (m FeatureMask) IsZero() bool { return m.bits[0] == 0 && m.bits[1] == 0 }

// Lo returns the low 64 bits.
func (m FeatureMask) Lo() uint64 { return m.bits[0] }

// Hi returns the high 64 bits.
func (m FeatureMask) Hi() uint64 { return m.bits[1] }

// Equal reports whether two masks are identical.
func (m FeatureMask) Equal(other FeatureMask) bool { return m.bits == other.bits }

// FromLo constructs a FeatureMask from the low 64 bits (hi=0). Used for
// testing/schema.
func FromLo(lo uint64) FeatureMask { return FeatureMask{bits: [2]uint64{lo, 0}} }

// Has reports whether bit position b is set (0-indexed). Positions outside the
// valid 0-127 range are reported as unset.
func (m FeatureMask) Has(b int) bool {
	if b < 0 || b > 127 {
		return false
	}
	if b < 64 {
		return m.bits[0]&(1<<uint(b)) != 0
	}
	return m.bits[1]&(1<<uint(b-64)) != 0
}

// FromBits constructs a FeatureMask from a list of bit positions (0-indexed).
// Positions outside the valid 0-127 range are ignored. This is the primary
// constructor used by the v0.1-en feature table (ADR-0008).
func FromBits(bits ...int) FeatureMask {
	var m FeatureMask
	for _, b := range bits {
		if b >= 0 && b < 64 {
			m.bits[0] |= 1 << uint(b)
		} else if b >= 64 && b < 128 {
			m.bits[1] |= 1 << uint(b-64)
		}
	}
	return m
}
