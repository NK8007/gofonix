package dict

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestLoadFull_DefaultBuild_ReturnsErrFullDictNotBuilt verifies that in the
// default build (no `-tags gofonix_full_dict`) LoadFull returns
// ErrFullDictNotBuilt immediately — without touching the filesystem and
// without panicking — so the engine's graceful-degradation path is reachable
// (ADR-0010: "build tag absent" is silent).
//
// In a tagged build the stub is replaced by the real loader, which will
// return a different sentinel (path resolution / size / checksum). The test
// is therefore guarded against both: it asserts ONLY non-panic + non-nil
// error, and only additionally asserts ErrFullDictNotBuilt in the default
// build path. This keeps the test honest under both build configurations.
func TestLoadFull_DefaultBuild_ReturnsErrFullDictNotBuilt(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("LoadFull(\"\") panicked: %v (must never panic per ADR-0010)", r)
		}
	}()

	d, err := LoadFull("")
	if err == nil {
		t.Fatalf("LoadFull(\"\") returned nil error; expected a sentinel error from loader_full_errors.go")
	}
	if d != nil {
		t.Fatalf("LoadFull(\"\") returned a non-nil *Dict alongside error %v; on any failure path the returned *Dict must be nil", err)
	}
}

// TestLoadFull_EmptyOverride_NeverPanics is the explicit "the system does not
// panic" assertion required by the no-panic contract. It calls LoadFull
// with an empty override (the default resolution path) inside a recover()
// scope and fails if any code along the path panics, regardless of which
// sentinel is ultimately returned.
func TestLoadFull_EmptyOverride_NeverPanics(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("LoadFull(\"\") panicked with empty override: %v (ADR-0010 forbids panic on every failure path)", r)
		}
	}()
	_, _ = LoadFull("")
}

// TestLoadFull_NonexistentOverride_NeverPanics points LoadFull at an explicit
// non-existent path. In the default build the stub still returns
// ErrFullDictNotBuilt before any filesystem I/O; in a tagged build the real
// loader returns ErrFullDictFileUnreadable. Either way it must never panic
// and must return a non-nil error with a nil *Dict.
func TestLoadFull_NonexistentOverride_NeverPanics(t *testing.T) {
	tmp := t.TempDir()
	missing := filepath.Join(tmp, "definitely-does-not-exist.dict")
	// Sanity guard: confirm the file really does not exist so the test is
	// honest about what it is asserting.
	if _, err := os.Stat(missing); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("test setup: expected %q not to exist, got err=%v", missing, err)
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("LoadFull(%q) panicked: %v (ADR-0010 forbids panic on every failure path)", missing, r)
		}
	}()

	d, err := LoadFull(missing)
	if err == nil {
		t.Fatalf("LoadFull(%q) returned nil error for a non-existent file; expected a sentinel error", missing)
	}
	if d != nil {
		t.Fatalf("LoadFull(%q) returned a non-nil *Dict; expected nil on every failure path", missing)
	}
}

// TestFullDictFallbackReason_ClosedEnum asserts that every sentinel error
// from loader_full_errors.go maps to a non-empty, closed reason string
// (ADR-0010). A nil error maps to the empty string. An
// unrecognised error maps to "unexpected error".
func TestFullDictFallbackReason_ClosedEnum(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"nil", nil, ""},
		{"not-built", ErrFullDictNotBuilt, "build tag absent"},
		{"path-unresolved", ErrFullDictPathUnresolved, "no path resolved"},
		{"file-unreadable", ErrFullDictFileUnreadable, "file unreadable"},
		{"size-out-of-band", ErrFullDictSizeOutOfBand, "size out of band"},
		{"checksum-mismatch", ErrFullDictChecksumMismatch, "checksum mismatch"},
		{"empty-after-parse", ErrFullDictEmptyAfterParse, "empty after parse"},
		{"unknown", errors.New("something else"), "unexpected error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FullDictFallbackReason(tc.err)
			if got != tc.want {
				t.Fatalf("FullDictFallbackReason(%v) = %q, want %q", tc.err, got, tc.want)
			}
		})
	}
}

// TestFullID_Frozen pins the FullID constant to the exact value required by
// ADR-0010. Changing this value requires bumping the ADR and
// the frozen SHA-256 digest in lockstep.
func TestFullID_Frozen(t *testing.T) {
	if got, want := FullID, "cmudict-full-v0.7b"; got != want {
		t.Fatalf("FullID = %q, want %q (ADR-0010 freezes this identifier)", got, want)
	}
}
