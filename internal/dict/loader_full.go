//go:build gofonix_full_dict

// Package-internal full-CMUdict loader, compiled in only when the binary is
// built with `-tags gofonix_full_dict` (ADR-0010, Decision §2). The default
// build path uses loader_full_stub.go instead, which returns ErrFullDictNotBuilt
// immediately so the engine falls back to the embedded mini-dict.
//
// The loader reads exactly one local file. It performs:
//   1. path resolution (GOFONIX_DICT_PATH → $HOME/.gofonix/cmudict.dict);
//   2. stat / regular-file check;
//   3. a size sanity band [1 MiB, 16 MiB] (ADR-0010, Loading Pipeline §3);
//   4. a streaming SHA-256 vs FullExpectedSHA256 (ADR-0010, Checksum Policy);
//   5. a parse via the existing Load(), which also runs in the mini path;
//   6. a "zero usable entries" rejection (ADR-0010, Loading Pipeline §5).
//
// Every step that declines returns one of the sentinel errors declared in
// loader_full_errors.go. The loader NEVER panics, NEVER touches the network,
// NEVER shells out, and NEVER mutates state outside the returned *Dict.

package dict

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	// fullDictMinSize and fullDictMaxSize bound the on-disk size of the full
	// CMUdict file (ADR-0010, Loading Pipeline §3). CMUdict v0.7b is ~3.5 MiB
	// in canonical form; the band catches obvious mistakes (empty file,
	// misrouted log, multi-gigabyte misconfiguration) without pretending to
	// be a security boundary. The SHA-256 check is the real authority.
	fullDictMinSize int64 = 1 << 20  // 1 MiB
	fullDictMaxSize int64 = 16 << 20 // 16 MiB

	// fullDictDefaultRelPath is the default location under $HOME for the full
	// CMUdict file when GOFONIX_DICT_PATH is unset (ADR-0010, Decision §3).
	fullDictDefaultRelPath = ".gofonix/cmudict.dict"

	// fullDictPathEnv is the environment variable that, when set and
	// non-empty, overrides the default path (ADR-0010, Decision §3).
	fullDictPathEnv = "GOFONIX_DICT_PATH"
)

// LoadFull loads the full CMUdict v0.7b from a local file and returns an
// immutable Dict tagged with FullID (ADR-0010). It is invoked once at engine
// construction time and is never retried.
//
// The override argument lets callers (notably the G2P engine and tests) pin
// a specific path; an empty override triggers the documented resolution
// order: GOFONIX_DICT_PATH → $HOME/.gofonix/cmudict.dict.
//
// On any failure path LoadFull returns a sentinel error from
// loader_full_errors.go (wrapped where useful via fmt.Errorf("%w: …", …)),
// never a panic. Successful return is the only path on which the caller
// should publish the dictionary as the engine's active backend.
func LoadFull(override string) (*Dict, error) {
	path, err := resolveFullDictPath(override)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrFullDictFileUnreadable, path, err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%w: %s: not a regular file", ErrFullDictFileUnreadable, path)
	}
	size := info.Size()
	if size < fullDictMinSize || size > fullDictMaxSize {
		return nil, fmt.Errorf("%w: %s: size=%d bytes", ErrFullDictSizeOutOfBand, path, size)
	}

	data, sum, err := readAndHashFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrFullDictFileUnreadable, path, err)
	}
	gotHex := hex.EncodeToString(sum)
	if gotHex != FullExpectedSHA256 {
		return nil, fmt.Errorf("%w: %s: got=%s want=%s", ErrFullDictChecksumMismatch, path, gotHex, FullExpectedSHA256)
	}

	d, err := Load(data)
	if err != nil {
		// Load is contractually non-erroring in Slice 2, but the signature
		// reserves an error return for forward compatibility (see dict.go).
		return nil, fmt.Errorf("%w: %s: %v", ErrFullDictEmptyAfterParse, path, err)
	}
	if d == nil || len(d.entries) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrFullDictEmptyAfterParse, path)
	}
	return d, nil
}

// resolveFullDictPath implements the path-resolution order from ADR-0010
// (Decision §3, Loading Pipeline §1):
//
//  1. an explicit non-empty override (callers may pin a path directly);
//  2. GOFONIX_DICT_PATH if set and non-empty;
//  3. $HOME/.gofonix/cmudict.dict if $HOME is set and non-empty;
//  4. otherwise return ErrFullDictPathUnresolved.
//
// The function performs only string-level resolution. No I/O. No path
// canonicalisation beyond filepath.Join (which is purely lexical for our
// inputs).
func resolveFullDictPath(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	if envPath := os.Getenv(fullDictPathEnv); envPath != "" {
		return envPath, nil
	}
	home := os.Getenv("HOME")
	if home == "" {
		return "", ErrFullDictPathUnresolved
	}
	return filepath.Join(home, fullDictDefaultRelPath), nil
}

// readAndHashFile streams the file at path through SHA-256 while accumulating
// its bytes for the parser. We compute the digest over the exact file bytes —
// no LF/CRLF normalisation, no whitespace trim, no comment stripping — to
// preserve the byte-equality property of the checksum (ADR-0010, Checksum
// Policy).
//
// The function uses io.TeeReader so a single pass produces both the digest
// and the payload, avoiding a second open or a second read of the file.
func readAndHashFile(path string) ([]byte, []byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	hasher := sha256.New()
	// Read into a fresh buffer via TeeReader so the parser sees the same
	// bytes the hasher hashed. Capping the read at fullDictMaxSize+1 lets us
	// detect a file that grew between Stat and Open beyond the sanity band.
	limited := io.LimitReader(io.TeeReader(f, hasher), fullDictMaxSize+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, nil, err
	}
	if int64(len(data)) > fullDictMaxSize {
		return nil, nil, errors.New("file grew past size band during read")
	}
	return data, hasher.Sum(nil), nil
}
