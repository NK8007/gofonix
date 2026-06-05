package phoneme

import "testing"

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
