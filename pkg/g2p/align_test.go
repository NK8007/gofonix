package g2p

import (
	"reflect"
	"testing"
)

// TestUniformByteAlignmentBasic pins the canonical 1:1 case from ADR-0007:
// "cat" with 3 phonemes over byte span [0,3) yields [0,1) [1,2) [2,3).
func TestUniformByteAlignmentBasic(t *testing.T) {
	got := uniformByteAlignment(0, 3, 3)
	want := []ByteSpan{{0, 1}, {1, 2}, {2, 3}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("uniformByteAlignment(0,3,3) = %v, want %v", got, want)
	}
}

// TestUniformByteAlignmentRemainderLeading checks that when M is not divisible
// by N, the first R = M mod N phonemes each receive one extra byte (the
// remainder goes to the leading phonemes). For M=7, N=3: base=2, R=1 -> widths
// 3,2,2.
func TestUniformByteAlignmentRemainderLeading(t *testing.T) {
	got := uniformByteAlignment(0, 7, 3)
	want := []ByteSpan{{0, 3}, {3, 5}, {5, 7}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("uniformByteAlignment(0,7,3) = %v, want %v", got, want)
	}
	// Widths: 3, 2, 2.
	widths := []int{3, 2, 2}
	for i, s := range got {
		if s.End-s.Start != widths[i] {
			t.Errorf("span %d width = %d, want %d", i, s.End-s.Start, widths[i])
		}
	}
}

// TestUniformByteAlignmentRemainderTwo checks a larger remainder: M=8, N=3 ->
// base=2, R=2 -> widths 3,3,2.
func TestUniformByteAlignmentRemainderTwo(t *testing.T) {
	got := uniformByteAlignment(0, 8, 3)
	want := []ByteSpan{{0, 3}, {3, 6}, {6, 8}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("uniformByteAlignment(0,8,3) = %v, want %v", got, want)
	}
}

// TestUniformByteAlignmentShortM exercises M < N: there are more phonemes than
// bytes, so some trailing spans are zero-width [k,k). For M=2, N=5: base=0,
// R=2 -> widths 1,1,0,0,0.
func TestUniformByteAlignmentShortM(t *testing.T) {
	got := uniformByteAlignment(0, 2, 5)
	want := []ByteSpan{{0, 1}, {1, 2}, {2, 2}, {2, 2}, {2, 2}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("uniformByteAlignment(0,2,5) = %v, want %v", got, want)
	}
	// Confirm the trailing three spans are zero-width.
	for i := 2; i < 5; i++ {
		if got[i].Start != got[i].End {
			t.Errorf("span %d = %v, want zero-width", i, got[i])
		}
	}
}

// TestUniformByteAlignmentZeroPhonemes confirms N==0 yields an empty alignment.
func TestUniformByteAlignmentZeroPhonemes(t *testing.T) {
	got := uniformByteAlignment(2, 5, 0)
	if len(got) != 0 {
		t.Errorf("uniformByteAlignment(2,5,0) = %v, want empty", got)
	}
}

// TestUniformByteAlignmentOffsetBase confirms the partition starts at w0 (not 0)
// and tiles [w0,w1) exactly.
func TestUniformByteAlignmentOffsetBase(t *testing.T) {
	got := uniformByteAlignment(10, 14, 2)
	want := []ByteSpan{{10, 12}, {12, 14}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("uniformByteAlignment(10,14,2) = %v, want %v", got, want)
	}
}

// TestUniformByteAlignmentPartitionInvariants verifies, across many (M,N)
// combinations, that the spans form a contiguous, non-overlapping partition of
// [w0,w1): the first span starts at w0, the last ends at w1, every span is
// non-negative width, and adjacent spans abut exactly.
func TestUniformByteAlignmentPartitionInvariants(t *testing.T) {
	const w0 = 5
	for m := 0; m <= 20; m++ {
		for n := 1; n <= 12; n++ {
			w1 := w0 + m
			spans := uniformByteAlignment(w0, w1, n)
			if len(spans) != n {
				t.Fatalf("M=%d N=%d: got %d spans, want %d", m, n, len(spans), n)
			}
			if spans[0].Start != w0 {
				t.Errorf("M=%d N=%d: first start = %d, want %d", m, n, spans[0].Start, w0)
			}
			if spans[n-1].End != w1 {
				t.Errorf("M=%d N=%d: last end = %d, want %d", m, n, spans[n-1].End, w1)
			}
			total := 0
			for i, s := range spans {
				if s.End < s.Start {
					t.Errorf("M=%d N=%d: span %d negative width %v", m, n, i, s)
				}
				if i > 0 && spans[i-1].End != s.Start {
					t.Errorf("M=%d N=%d: span %d not contiguous: prev end %d, start %d",
						m, n, i, spans[i-1].End, s.Start)
				}
				total += s.End - s.Start
			}
			if total != m {
				t.Errorf("M=%d N=%d: total width = %d, want %d", m, n, total, m)
			}
			// Remainder property: the first R spans are widest. For m >= n,
			// every span width is base or base+1, and exactly R = m mod n are
			// base+1, all leading.
			base := m / n
			rem := m % n
			for i, s := range spans {
				w := s.End - s.Start
				wantW := base
				if i < rem {
					wantW = base + 1
				}
				if w != wantW {
					t.Errorf("M=%d N=%d: span %d width = %d, want %d", m, n, i, w, wantW)
				}
			}
		}
	}
}
