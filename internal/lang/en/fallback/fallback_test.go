package fallback

import (
	"reflect"
	"testing"
)

// --- A. Rule table validation ---

func TestValidateRules(t *testing.T) {
	if err := validateRules(); err != nil {
		t.Fatalf("validateRules() returned error: %v", err)
	}
}

func TestRulesVersion(t *testing.T) {
	if RulesVersion != "fallback-en-v0.2" {
		t.Fatalf("RulesVersion = %q, want %q", RulesVersion, "fallback-en-v0.2")
	}
}

func TestXEmitsTwoSymbols(t *testing.T) {
	got, ok := ruleIndex["x"]
	if !ok {
		t.Fatalf("rule for %q not found", "x")
	}
	want := []string{"K", "S"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("x -> %v, want %v", got, want)
	}
}

func TestThMapsToDH(t *testing.T) {
	got, ok := ruleIndex["th"]
	if !ok {
		t.Fatalf("rule for %q not found", "th")
	}
	want := []string{"DH"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("th -> %v, want %v (must be DH, not TH)", got, want)
	}
}

func TestAllSingleLettersPresent(t *testing.T) {
	for c := byte('a'); c <= 'z'; c++ {
		if _, ok := ruleIndex[string(c)]; !ok {
			t.Errorf("missing single-letter rule for %q", string(c))
		}
	}
}

// --- B. Normalization ---

func TestNormalize(t *testing.T) {
	cases := []struct {
		name      string
		input     string
		want      string
		wantValid bool
	}{
		{"lowercase passthrough", "qwerty", "qwerty", true},
		{"uppercase folds", "Qwerty", "qwerty", true},
		{"ascii apostrophe stripped", "don't", "dont", true},
		{"multiple ascii apostrophes", "rock'n'roll", "rocknroll", true},
		{"digits rejected MP3", "MP3", "", false},
		{"digits rejected H2O", "H2O", "", false},
		{"digits rejected 3rd", "3rd", "", false},
		{"non-ascii rejected cafe", "café", "", false},
		{"curly apostrophe rejected", "it\u2019s", "", false},
		{"empty input", "", "", false},
		{"apostrophe only becomes empty", "'", "", false},

		// ASCII uppercase — must still work after strict ASCII fold.
		{"ascii single uppercase K", "K", "k", true},
		{"ascii all uppercase HELLO", "HELLO", "hello", true},
		{"ascii uppercase with apostrophe DON'T", "DON'T", "dont", true},

		// Non-ASCII — must be declined (regression for ADR-0009 ASCII-only
		// activation predicate; strings.ToLower could have leaked these).
		{"turkish capital I with dot U+0130", "İ", "", false},
		{"n with tilde U+00F1", "ñ", "", false},
		{"A with umlaut U+00C4", "Ä", "", false},
		{"naive with i diaeresis U+00EF", "naïve", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := Normalize(tc.input)
			if got != tc.want || ok != tc.wantValid {
				t.Fatalf("Normalize(%q) = (%q, %v), want (%q, %v)", tc.input, got, ok, tc.want, tc.wantValid)
			}
		})
	}
}

// --- C. Longest-match (frozen table) ---

func TestMatch(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"ship", "ship", []string{"SH", "IH", "P"}},
		{"chat", "chat", []string{"CH", "AE", "T"}},
		{"thing", "thing", []string{"DH", "IH", "NG"}},
		{"phone", "phone", []string{"F", "AO", "N", "EH"}},
		{"see", "see", []string{"S", "IY"}},
		{"quick", "quick", []string{"K", "AH", "IH", "K"}},
		{"dont", "dont", []string{"D", "AO", "N", "T"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := Match(tc.in)
			if !ok {
				t.Fatalf("Match(%q) declined, want %v", tc.in, tc.want)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Match(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

// TestMatchNonLetterDeclines pins that Match declines (does not panic) when a
// non-letter byte that would have been rejected by Normalize is fed directly.
func TestMatchNonLetterDeclines(t *testing.T) {
	got, ok := Match("!")
	if ok || got != nil {
		t.Fatalf("Match(%q) = (%v, %v), want (nil, false)", "!", got, ok)
	}
}

func TestMatchXenonStartsWithKS(t *testing.T) {
	got, ok := Match("xenon")
	if !ok {
		t.Fatalf("Match(\"xenon\") declined")
	}
	if len(got) < 2 || got[0] != "K" || got[1] != "S" {
		t.Fatalf("Match(\"xenon\") = %v, want prefix [K S]", got)
	}
}

func TestMatchEmptyDeclines(t *testing.T) {
	if got, ok := Match(""); ok || got != nil {
		t.Fatalf("Match(\"\") = (%v, %v), want (nil, false)", got, ok)
	}
}

// --- D. Determinism ---

func TestPronounceDeterministic(t *testing.T) {
	first, ok := Pronounce("hello")
	if !ok {
		t.Fatalf("Pronounce(\"hello\") declined")
	}
	for n := 0; n < 5; n++ {
		got, ok := Pronounce("hello")
		if !ok {
			t.Fatalf("Pronounce(\"hello\") declined on iteration %d", n)
		}
		if !reflect.DeepEqual(got, first) {
			t.Fatalf("Pronounce(\"hello\") iteration %d = %v, want %v", n, got, first)
		}
	}
}

func TestPronounceDeclinesIneligible(t *testing.T) {
	for _, in := range []string{"MP3", "H2O", "3rd", "café", "", "'"} {
		if got, ok := Pronounce(in); ok || got != nil {
			t.Errorf("Pronounce(%q) = (%v, %v), want (nil, false)", in, got, ok)
		}
	}
}

// Every RHS symbol must be mappable by the existing ARPAbet bridge would be an
// integration concern; here we keep validation closed within the package via
// validateRules (which checks membership in the v0.1-en inventory).
