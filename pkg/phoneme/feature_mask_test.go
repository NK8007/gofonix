package phoneme

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestFeatureMaskMarshalJSON pins the JSON serialization of FeatureMask to the
// {"lo":...,"hi":...} feature-word form required by the golden JSONL schema
// (docs/golden_jsonl_schema.md). It is a regression guard against the prior
// behaviour where the unexported bits field caused encoding/json to emit an
// empty object {}, silently dropping every feature value from a marshalled
// Result (notably the CLI's --output json FeatureStream).
func TestFeatureMaskMarshalJSON(t *testing.T) {
	// A representative non-zero mask using both the low and high words so the
	// test would fail if either word were dropped or swapped.
	m := FromBits(2, 10, 64) // lo has bits 2 and 10 set (0x404 = 1028), hi has bit 0 (1)
	wantLo := uint64(1<<2 | 1<<10)
	wantHi := uint64(1)
	if m.Lo() != wantLo || m.Hi() != wantHi {
		t.Fatalf("test setup wrong: Lo/Hi = %d/%d, want %d/%d", m.Lo(), m.Hi(), wantLo, wantHi)
	}

	b, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("json.Marshal(FeatureMask) failed: %v", err)
	}
	got := string(b)

	// Hard regression check: the mask MUST NOT serialize as an empty object.
	if got == "{}" {
		t.Fatalf("FeatureMask serialized as empty object %q; feature values were dropped", got)
	}
	if !strings.Contains(got, "\"lo\":") || !strings.Contains(got, "\"hi\":") {
		t.Fatalf("FeatureMask JSON %q missing \"lo\"/\"hi\" keys; want {lo,hi} form", got)
	}

	// Round-trip the JSON into a neutral {lo,hi} struct and confirm the values
	// match the accessors exactly (lossless).
	var back struct {
		Lo uint64 `json:"lo"`
		Hi uint64 `json:"hi"`
	}
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("unmarshal of FeatureMask JSON %q failed: %v", got, err)
	}
	if back.Lo != wantLo || back.Hi != wantHi {
		t.Errorf("round-trip lo/hi = %d/%d, want %d/%d (json=%q)", back.Lo, back.Hi, wantLo, wantHi, got)
	}

	// The zero mask must still serialize with explicit zero words, never {}.
	zb, err := json.Marshal(Zero())
	if err != nil {
		t.Fatalf("json.Marshal(Zero()) failed: %v", err)
	}
	if string(zb) == "{}" {
		t.Fatalf("Zero() serialized as empty object %q; want explicit {\"lo\":0,\"hi\":0}", string(zb))
	}
	if string(zb) != `{"lo":0,"hi":0}` {
		t.Errorf("Zero() JSON = %q, want {\"lo\":0,\"hi\":0}", string(zb))
	}
}

func TestZeroIsZero(t *testing.T) {
	if !Zero().IsZero() {
		t.Fatalf("Zero().IsZero() = false, want true")
	}
	if Zero().Lo() != 0 || Zero().Hi() != 0 {
		t.Fatalf("Zero() lo/hi = %d/%d, want 0/0", Zero().Lo(), Zero().Hi())
	}
}

func TestFromLo(t *testing.T) {
	m := FromLo(265)
	if m.Lo() != 265 {
		t.Errorf("FromLo(265).Lo() = %d, want 265", m.Lo())
	}
	if m.Hi() != 0 {
		t.Errorf("FromLo(265).Hi() = %d, want 0", m.Hi())
	}
	if m.IsZero() {
		t.Errorf("FromLo(265).IsZero() = true, want false")
	}
}

func TestEqual(t *testing.T) {
	if !Zero().Equal(Zero()) {
		t.Errorf("Zero().Equal(Zero()) = false, want true")
	}
	if FromLo(1).Equal(FromLo(2)) {
		t.Errorf("FromLo(1).Equal(FromLo(2)) = true, want false")
	}
	if !FromLo(265).Equal(FromLo(265)) {
		t.Errorf("FromLo(265).Equal(FromLo(265)) = false, want true")
	}
}

// TestBoundaryMaskVsZero pins the ADR-0008 disambiguation: the boundary phoneme
// ID 0 carries lo=4194304 (bit 22), which is NOT the zero mask.
func TestBoundaryMaskVsZero(t *testing.T) {
	boundary := FromLo(4194304) // bit 22 set
	if boundary.IsZero() {
		t.Errorf("boundary mask (lo=4194304) reports IsZero, want non-zero")
	}
	if boundary.Equal(Zero()) {
		t.Errorf("boundary mask must not equal the zero mask")
	}
}

func TestFromBitsAndHas(t *testing.T) {
	m := FromBits(0, 14, 16, 21) // AA = lo 2179073
	if m.Lo() != 2179073 {
		t.Errorf("FromBits(0,14,16,21).Lo() = %d, want 2179073", m.Lo())
	}
	for _, b := range []int{0, 14, 16, 21} {
		if !m.Has(b) {
			t.Errorf("Has(%d) = false, want true", b)
		}
	}
	for _, b := range []int{1, 13, 15, 22, -1, 200} {
		if m.Has(b) {
			t.Errorf("Has(%d) = true, want false", b)
		}
	}
	// High-word bit.
	hi := FromBits(64)
	if hi.Hi() != 1 {
		t.Errorf("FromBits(64).Hi() = %d, want 1", hi.Hi())
	}
	if !hi.Has(64) {
		t.Error("FromBits(64).Has(64) = false, want true")
	}
	// Out-of-range bits are ignored.
	if !FromBits(-1, 128, 999).IsZero() {
		t.Error("FromBits with only out-of-range bits should be zero")
	}
}
