package golden_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/NK8007/gofonix/internal/golden"
	"github.com/NK8007/gofonix/pkg/g2p"
)

// TestModeRoundTrip confirms ModeString/ParseMode are mutually inverse for
// every valid mode and that an unknown spelling is rejected.
func TestModeRoundTrip(t *testing.T) {
	for _, m := range []g2p.Mode{g2p.ModeBatch, g2p.ModeOracle, g2p.ModeCausal} {
		s := golden.ModeString(m)
		got, ok := golden.ParseMode(s)
		if !ok || got != m {
			t.Errorf("ParseMode(ModeString(%v)) = (%v,%v), want (%v,true)", m, got, ok, m)
		}
	}
	if _, ok := golden.ParseMode("streaming"); ok {
		t.Error("ParseMode(\"streaming\") = ok, want not ok")
	}
}

// TestFromResultAndWriteRead exercises FromResult -> WriteAll -> ReadAll and
// confirms the key fields survive a JSONL round trip.
func TestFromResultAndWriteRead(t *testing.T) {
	e, err := g2p.New(g2p.Options{})
	if err != nil {
		t.Fatal(err)
	}
	res, err := e.Process("cat")
	if err != nil {
		t.Fatal(err)
	}
	rec := golden.FromResult(res)
	if rec.Input != "cat" || rec.Mode != "batch" {
		t.Fatalf("unexpected record header: %+v", rec)
	}
	if len(rec.FeatureStream) != 3 {
		t.Fatalf("feature_stream len = %d, want 3", len(rec.FeatureStream))
	}
	if rec.FeatureStream[0].Lo != 1028 {
		t.Errorf("feature_stream[0].lo = %d, want 1028 (K)", rec.FeatureStream[0].Lo)
	}
	if len(rec.Tokens) != 1 || rec.Tokens[0].Source != "dict" {
		t.Fatalf("unexpected tokens: %+v", rec.Tokens)
	}

	var buf bytes.Buffer
	if err := golden.WriteAll(&buf, []golden.Record{rec}); err != nil {
		t.Fatal(err)
	}
	// Exactly one trailing newline-terminated line.
	if n := bytes.Count(buf.Bytes(), []byte{'\n'}); n != 1 {
		t.Errorf("WriteAll wrote %d newlines, want 1", n)
	}

	back, err := golden.ReadAll(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if len(back) != 1 {
		t.Fatalf("ReadAll returned %d records, want 1", len(back))
	}
	if back[0].Input != "cat" || back[0].Tokens[0].Phonemes[0].ID != 20 {
		t.Errorf("round-trip mismatch: %+v", back[0])
	}
}

// TestReadAllSkipsBlankLines confirms blank lines are ignored and a malformed
// line is reported with its 1-based line number.
func TestReadAllSkipsBlankLines(t *testing.T) {
	good := `{"schema_version":"gofonix-result-v0.1","input":"x","mode":"batch"}`
	in := "\n" + good + "\n\n   \n"
	recs, err := golden.ReadAll(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 {
		t.Fatalf("ReadAll returned %d records, want 1", len(recs))
	}

	bad := good + "\nnot json\n"
	if _, err := golden.ReadAll(strings.NewReader(bad)); err == nil {
		t.Error("ReadAll on malformed line returned nil error, want error")
	} else if !strings.Contains(err.Error(), "line 2") {
		t.Errorf("error %q does not name the offending line", err)
	}
}
