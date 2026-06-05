//go:build gptreview
// +build gptreview

package fallback

import "testing"

func TestGPTReviewUnicodeFold(t *testing.T) {
	for _, in := range []string{"K", "İ"} {
		got, ok := Normalize(in)
		t.Logf("Normalize(%q) = (%q, %v)", in, got, ok)
	}
}
