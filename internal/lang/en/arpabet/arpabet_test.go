package arpabet

import (
	"testing"

	"github.com/gofonix/gofonix/pkg/phoneme"
)

func TestStripStress(t *testing.T) {
	cases := map[string]string{
		"AE0": "AE",
		"AE1": "AE",
		"AE2": "AE",
		"B":   "B",
		"HH":  "HH",
		"IH0": "IH",
		"":    "",
	}
	for in, want := range cases {
		if got := StripStress(in); got != want {
			t.Errorf("StripStress(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMapSymbol(t *testing.T) {
	cases := map[string]int{
		"AE":  2,
		"AE1": 2, // stress stripped
		"B":   7,
		"ZH":  39,
		"k":   20, // lower-case upper-cased
	}
	for in, wantID := range cases {
		p, ok := MapSymbol(in)
		if !ok {
			t.Errorf("MapSymbol(%q) ok = false, want true", in)
			continue
		}
		if p.ID != wantID {
			t.Errorf("MapSymbol(%q).ID = %d, want %d", in, p.ID, wantID)
		}
	}
	// Unknown symbol.
	if p, ok := MapSymbol("ZZZZ"); ok {
		t.Errorf("MapSymbol(\"ZZZZ\") = (%+v, true), want (Phoneme{}, false)", p)
	}
}

func TestMapSequence(t *testing.T) {
	got, ok := MapSequence([]string{"K", "AE1", "T"})
	if !ok {
		t.Fatal("MapSequence ok = false, want true")
	}
	wantIDs := []int{20, 2, 31}
	if len(got) != len(wantIDs) {
		t.Fatalf("len = %d, want %d", len(got), len(wantIDs))
	}
	for i, id := range wantIDs {
		if got[i].ID != id {
			t.Errorf("phoneme[%d].ID = %d, want %d", i, got[i].ID, id)
		}
	}
}

func TestMapSequenceUnknown(t *testing.T) {
	got, ok := MapSequence([]string{"K", "ZZZZ", "T"})
	if ok {
		t.Errorf("MapSequence ok = true, want false")
	}
	if got != nil {
		t.Errorf("MapSequence result = %v, want nil", got)
	}
}

// featureCases pins the ADR-0008 v0.1-en FeatureMask.lo decimals.
func TestFeatureMaskAA(t *testing.T) { assertLo(t, 1, 2179073) }
func TestFeatureMaskAE(t *testing.T) { assertLo(t, 2, 2146305) }
func TestFeatureMaskAH(t *testing.T) { assertLo(t, 3, 2105345) }
func TestFeatureMaskDH(t *testing.T) { assertLo(t, 10, 265) }
func TestFeatureMaskEY(t *testing.T) { assertLo(t, 13, 2924545) }
func TestFeatureMaskW(t *testing.T)  { assertLo(t, 36, 1185) }

// The following pin the Slice 4.1 ADR-0008 corrections so the previously wrong
// values (AO=2310145, D=261, G=1029) and the IH/M ID misassignment can never
// regress.
func TestFeatureMaskAO(t *testing.T) { assertLo(t, 4, 2113541) }
func TestFeatureMaskD(t *testing.T)  { assertLo(t, 9, 9) }
func TestFeatureMaskG(t *testing.T)  { assertLo(t, 15, 5) }

// TestFeatureMaskIH pins IH at ID 22 with lo=2105345 (bits 0,13,21), which is
// deliberately identical to AH (ID 3) in v0.1-en per ADR-0008.
func TestFeatureMaskIH(t *testing.T) {
	assertLo(t, 22, 2105345)
	if !FeatureFor(22).Equal(FeatureFor(3)) {
		t.Errorf("IH (ID 22) mask must equal AH (ID 3) mask in v0.1-en")
	}
}

// TestFeatureMaskM pins M at ID 17 (lo=131) after the IH/M swap.
func TestFeatureMaskM(t *testing.T) { assertLo(t, 17, 131) }

// TestMapSymbolIHM pins the corrected symbol->ID assignment: IH->22, M->17.
func TestMapSymbolIHM(t *testing.T) {
	cases := map[string]int{
		"IH":  22,
		"IH0": 22, // stress stripped
		"IH1": 22,
		"M":   17,
	}
	for in, wantID := range cases {
		p, ok := MapSymbol(in)
		if !ok {
			t.Errorf("MapSymbol(%q) ok = false, want true", in)
			continue
		}
		if p.ID != wantID {
			t.Errorf("MapSymbol(%q).ID = %d, want %d", in, p.ID, wantID)
		}
	}
}

// TestMapSequenceDOG pins DOG -> D AO G -> IDs [9, 4, 15] (Slice 4.1 fix).
func TestMapSequenceDOG(t *testing.T) {
	got, ok := MapSequence([]string{"D", "AO1", "G"})
	if !ok {
		t.Fatal("MapSequence(DOG) ok = false, want true")
	}
	wantIDs := []int{9, 4, 15}
	if len(got) != len(wantIDs) {
		t.Fatalf("len = %d, want %d", len(got), len(wantIDs))
	}
	for i, id := range wantIDs {
		if got[i].ID != id {
			t.Errorf("DOG phoneme[%d].ID = %d, want %d", i, got[i].ID, id)
		}
	}
}

// TestMapSequenceIT pins IT -> IH T -> IDs [22, 31] (Slice 4.1 IH fix).
func TestMapSequenceIT(t *testing.T) {
	got, ok := MapSequence([]string{"IH1", "T"})
	if !ok {
		t.Fatal("MapSequence(IT) ok = false, want true")
	}
	wantIDs := []int{22, 31}
	if len(got) != len(wantIDs) {
		t.Fatalf("len = %d, want %d", len(got), len(wantIDs))
	}
	for i, id := range wantIDs {
		if got[i].ID != id {
			t.Errorf("IT phoneme[%d].ID = %d, want %d", i, got[i].ID, id)
		}
	}
}

func assertLo(t *testing.T, id int, want uint64) {
	t.Helper()
	if got := FeatureFor(id).Lo(); got != want {
		t.Errorf("FeatureFor(%d).Lo() = %d, want %d", id, got, want)
	}
}

// TestAllFeatureMasks exhaustively pins every v0.1-en phoneme ID (0-39) to its
// ADR-0008 lo decimal, so any accidental bit-table edit is caught.
func TestAllFeatureMasks(t *testing.T) {
	want := map[int]uint64{
		0: 4194304, 1: 2179073, 2: 2146305, 3: 2105345, 4: 2113541, 5: 2703361,
		6: 2670593, 7: 133, 8: 528, 9: 9, 10: 265, 11: 2138113, 12: 3153921,
		13: 2924545, 14: 136, 15: 5, 16: 2056, 17: 131, 18: 2396161,
		19: 529, 20: 1028, 21: 353, 22: 2105345, 23: 259, 24: 1027, 25: 2826241,
		26: 2826241, 27: 132, 28: 289, 29: 264, 30: 520, 31: 260, 32: 264,
		33: 2297857, 34: 2560001, 35: 137, 36: 1185, 37: 545, 38: 265, 39: 521,
	}
	for id, w := range want {
		if got := FeatureFor(id).Lo(); got != w {
			t.Errorf("FeatureFor(%d).Lo() = %d, want %d", id, got, w)
		}
		if FeatureFor(id).Hi() != 0 {
			t.Errorf("FeatureFor(%d).Hi() = %d, want 0 (v0.1-en uses lo word only)", id, FeatureFor(id).Hi())
		}
	}
}

// TestBoundaryMask pins the boundary phoneme (ID 0): lo=4194304 (bit 22) and
// NOT the zero mask (ADR-0007/ADR-0008 disambiguation).
func TestBoundaryMask(t *testing.T) {
	m := FeatureFor(0)
	if m.Lo() != 4194304 {
		t.Errorf("FeatureFor(0).Lo() = %d, want 4194304", m.Lo())
	}
	if m.IsZero() {
		t.Errorf("FeatureFor(0).IsZero() = true, want false (boundary is non-zero)")
	}
}

// TestZeroVsBoundary pins the critical distinction: the zero mask is zero, the
// boundary phoneme mask is not.
func TestZeroVsBoundary(t *testing.T) {
	if !phoneme.Zero().IsZero() {
		t.Error("Zero().IsZero() = false, want true")
	}
	if FeatureFor(0).IsZero() {
		t.Error("FeatureFor(0).IsZero() = true, want false")
	}
	if FeatureFor(0).Equal(phoneme.Zero()) {
		t.Error("boundary mask equals zero mask, want distinct")
	}
}

// TestFeatureForUnknownID confirms an out-of-inventory ID returns the zero mask.
func TestFeatureForUnknownID(t *testing.T) {
	if !FeatureFor(9999).IsZero() {
		t.Error("FeatureFor(9999).IsZero() = false, want true (unknown ID -> zero mask)")
	}
}
