package dict

import (
	"reflect"
	"testing"
)

// --- Parser tests (ADR-0004) ---

func TestParserComments(t *testing.T) {
	for _, line := range []string{
		";;; CMUdict mini fixture v0.1 — for testing only",
		";;; Source: selected entries",
		";;;",
	} {
		if _, _, _, ok := parseLine(line); ok {
			t.Errorf("parseLine(%q) ok = true, want false (comment must be skipped)", line)
		}
	}
}

func TestParserSimpleEntry(t *testing.T) {
	key, phones, variant, ok := parseLine("CAT  K AE1 T")
	if !ok {
		t.Fatalf("parseLine ok = false, want true")
	}
	if key != "cat" {
		t.Errorf("key = %q, want %q", key, "cat")
	}
	if variant != 0 {
		t.Errorf("variant = %d, want 0", variant)
	}
	want := []string{"K", "AE1", "T"}
	if !reflect.DeepEqual(phones, want) {
		t.Errorf("phones = %v, want %v", phones, want)
	}
}

func TestParserAlternate(t *testing.T) {
	key, phones, variant, ok := parseLine("A(1)  EY1")
	if !ok {
		t.Fatalf("parseLine ok = false, want true")
	}
	if key != "a" {
		t.Errorf("key = %q, want %q", key, "a")
	}
	if variant != 1 {
		t.Errorf("variant = %d, want 1", variant)
	}
	want := []string{"EY1"}
	if !reflect.DeepEqual(phones, want) {
		t.Errorf("phones = %v, want %v", phones, want)
	}
}

func TestParserEmptyLine(t *testing.T) {
	for _, line := range []string{"", "   ", "\t", "  \t "} {
		if _, _, _, ok := parseLine(line); ok {
			t.Errorf("parseLine(%q) ok = true, want false (empty line must be skipped)", line)
		}
	}
}

func TestParserNoPhones(t *testing.T) {
	for _, line := range []string{"CAT", "CAT  ", "DOG"} {
		if _, _, _, ok := parseLine(line); ok {
			t.Errorf("parseLine(%q) ok = true, want false (no phonemes must be skipped)", line)
		}
	}
}

// --- Lookup tests (ADR-0004, ADR-0006) ---

func loadTestDict(t *testing.T) *Dict {
	t.Helper()
	d, err := LoadEmbedded()
	if err != nil {
		t.Fatalf("LoadEmbedded() error = %v", err)
	}
	return d
}

func TestLookupCanonical(t *testing.T) {
	d := loadTestDict(t)
	e, ok := d.Lookup("cat")
	if !ok {
		t.Fatal("Lookup(\"cat\") not found, want found")
	}
	want := []string{"K", "AE1", "T"}
	if !reflect.DeepEqual(e.Canonical, want) {
		t.Errorf("Canonical = %v, want %v", e.Canonical, want)
	}
}

func TestLookupCaseInsensitive(t *testing.T) {
	d := loadTestDict(t)
	eLower, okLower := d.Lookup("cat")
	eUpper, okUpper := d.Lookup("CAT")
	eMixed, okMixed := d.Lookup("Cat")
	if !okLower || !okUpper || !okMixed {
		t.Fatalf("found = (%v,%v,%v), want all true", okLower, okUpper, okMixed)
	}
	if !reflect.DeepEqual(eLower, eUpper) || !reflect.DeepEqual(eLower, eMixed) {
		t.Errorf("case-insensitive lookups differ: %v / %v / %v", eLower, eUpper, eMixed)
	}
}

func TestLookupAlternateOnly(t *testing.T) {
	d := loadTestDict(t)
	// "ONLY(1)" exists but there is no unnumbered "ONLY" base entry.
	if e, ok := d.Lookup("only"); ok {
		t.Errorf("Lookup(\"only\") found = true (Canonical=%v), want not found", e.Canonical)
	}
}

func TestLookupAlternateNotSelected(t *testing.T) {
	d := loadTestDict(t)
	// "A" has unnumbered AH0 and alternate A(1) EY1. Canonical must be AH0.
	e, ok := d.Lookup("a")
	if !ok {
		t.Fatal("Lookup(\"a\") not found, want found")
	}
	want := []string{"AH0"}
	if !reflect.DeepEqual(e.Canonical, want) {
		t.Errorf("Canonical = %v, want %v (alternate EY1 must NOT be selected)", e.Canonical, want)
	}
	// The alternate is still stored for future use.
	wantAlt := [][]string{{"EY1"}}
	if !reflect.DeepEqual(e.Alternates, wantAlt) {
		t.Errorf("Alternates = %v, want %v", e.Alternates, wantAlt)
	}
}

func TestLookupMissing(t *testing.T) {
	d := loadTestDict(t)
	if e, ok := d.Lookup("xyzzy"); ok {
		t.Errorf("Lookup(\"xyzzy\") found = true (Canonical=%v), want not found", e.Canonical)
	}
}

// TestLookupMutationIsolation verifies that mutating a returned Entry
// does not corrupt subsequent Lookup results (Dict immutability).
func TestLookupMutationIsolation(t *testing.T) {
	d, err := LoadEmbedded()
	if err != nil {
		t.Fatalf("LoadEmbedded: %v", err)
	}
	e1, ok := d.Lookup("cat")
	if !ok {
		t.Fatal("cat not found")
	}
	original := e1.Canonical[0]
	e1.Canonical[0] = "MUTATED" // mutate the returned copy

	e2, ok := d.Lookup("cat")
	if !ok {
		t.Fatal("cat not found on second lookup")
	}
	if e2.Canonical[0] != original {
		t.Errorf("Dict mutated: got %q, want %q", e2.Canonical[0], original)
	}
}

// TestParserMalformedVariantSuffix checks that malformed variant suffixes are
// skipped by the parser. Per the ADR-0004 parser (parser.go): an empty interior,
// a zero or non-positive variant number, a non-numeric interior, and an unclosed
// paren are all treated as malformed and skipped (ok=false). Note variant 0 is
// NOT valid as an explicit "(0)" suffix — the canonical entry is the *unnumbered*
// form, and an explicit "(0)" fails the n < 1 check and is skipped.
func TestParserMalformedVariantSuffix(t *testing.T) {
	cases := []string{
		"WORD()  K AE1 T",  // empty number
		"WORD(0)  K AE1 T", // zero not valid alternate
		"WORD(X)  K AE1 T", // non-numeric
		"WORD(1   K AE1 T", // unclosed paren
	}
	for _, line := range cases {
		_, _, _, ok := parseLine(line)
		if ok {
			t.Errorf("expected parseLine(%q) to be skipped, got ok=true", line)
		}
	}
}

// --- Embed + checksum tests (ADR-0004) ---

func TestLoadEmbedded(t *testing.T) {
	if _, err := LoadEmbedded(); err != nil {
		t.Fatalf("LoadEmbedded() error = %v, want nil", err)
	}
}

func TestEmbeddedChecksum(t *testing.T) {
	c1 := EmbeddedChecksum()
	if c1 == "" {
		t.Fatal("EmbeddedChecksum() is empty, want non-empty")
	}
	c2 := EmbeddedChecksum()
	if c1 != c2 {
		t.Errorf("EmbeddedChecksum() not stable: %q != %q", c1, c2)
	}
	// SHA-256 hex is 64 characters.
	if len(c1) != 64 {
		t.Errorf("EmbeddedChecksum() length = %d, want 64", len(c1))
	}
}
