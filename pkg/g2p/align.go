package g2p

// uniformByteAlignment partitions the byte span [w0, w1) into n contiguous,
// non-overlapping ByteSpans, one per phoneme, per the uniform byte-based
// alignment of ADR-0007.
//
// The algorithm is purely byte-based and is intentionally NOT rune-aware: a
// span boundary may fall inside a multi-byte UTF-8 code point. This is by
// design — alignment is a byte-offset projection device, not a grapheme
// segmentation.
//
// Given M = w1 - w0 input bytes and n phonemes:
//
//   - if n == 0, the alignment is empty;
//   - otherwise the n spans tile [w0, w1) exactly (their union is [w0, w1)
//     with no gaps and no overlaps);
//   - the first R = M mod n phonemes each receive floor(M/n)+1 bytes;
//   - the remaining n-R phonemes each receive floor(M/n) bytes;
//   - when M < n, the trailing spans are zero-width [k, k).
//
// Precondition: 0 <= w0 <= w1. The caller (Process) always passes a valid token
// span, so M is non-negative.
func uniformByteAlignment(w0, w1, n int) []ByteSpan {
	if n <= 0 {
		return []ByteSpan{}
	}
	m := w1 - w0
	if m < 0 {
		m = 0
	}
	base := m / n // floor(M/n)
	rem := m % n  // R = M mod n; the first R phonemes get one extra byte
	spans := make([]ByteSpan, n)
	pos := w0
	for i := 0; i < n; i++ {
		width := base
		if i < rem {
			width++
		}
		spans[i] = ByteSpan{Start: pos, End: pos + width}
		pos += width
	}
	// Postcondition: pos == w1 (the partition exactly tiles [w0, w1)).
	return spans
}
