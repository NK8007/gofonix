//go:build unicodefold
// +build unicodefold

package fallback

import "testing"

func TestNormalizeUnicodeFold(t *testing.T) {
	for _, in := range []string{"K", "İ"} {
		got, ok := Normalize(in)
		t.Logf("Normalize(%q) = (%q, %v)", in, got, ok)
	}
}
