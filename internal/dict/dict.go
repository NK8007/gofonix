// Package dict implements the internal mini CMUdict infrastructure for v0.1
// (ADR-0004): a deterministic CMUdict-format parser, an embedded mini fixture,
// a case-insensitive lookup, and SHA-256 provenance metadata.
//
// This is an internal package and carries no API stability guarantee. The
// public g2p package consumes it to populate Result.Metadata. This package
// stores RAW ARPAbet symbols only — mapping to neutral phonemes and stress
// stripping are performed by the internal/lang/en/arpabet bridge, not here.
package dict

import (
	"sort"
	"strings"
)

// Dict is an immutable, in-memory CMUdict (ADR-0004). It is constructed once by
// Load (or LoadEmbedded) and is safe for concurrent Lookup calls thereafter.
type Dict struct {
	// entries maps a lowercased lookup key to its Entry. Only keys that have an
	// unnumbered (canonical) pronunciation are exposed by Lookup; keys that have
	// only numbered alternates are deliberately not surfaced (see Lookup).
	entries map[string]*Entry
}

// pendingEntry accumulates parsed pronunciations for a key before the final
// Entry is materialized. variant 0 fills canonical; variants >= 1 are collected
// with their numbers so they can be sorted.
type pendingEntry struct {
	canonical    []string
	hasCanonical bool
	alternates   []variantPhones
}

// variantPhones pairs an alternate's variant number with its phoneme sequence.
type variantPhones struct {
	variant int
	phones  []string
}

// Load parses CMUdict-format data and returns an immutable Dict (ADR-0004).
//
// All lines are parsed deterministically. Comment lines (";;;"), blank lines,
// and malformed lines are skipped. The unnumbered (variant 0) pronunciation of
// a key becomes Entry.Canonical; numbered alternates (variant >= 1) are appended
// to Entry.Alternates in ascending variant order. Load never returns an error
// today (malformed lines are skipped, not rejected), but the error return
// is part of the contract for forward compatibility.
func Load(data []byte) (*Dict, error) {
	pending := make(map[string]*pendingEntry)

	// Iterate lines without allocating a full slice of all lines.
	s := string(data)
	for len(s) > 0 {
		var line string
		if idx := strings.IndexByte(s, '\n'); idx >= 0 {
			line = s[:idx]
			s = s[idx+1:]
		} else {
			line = s
			s = ""
		}
		// Strip a trailing CR for CRLF inputs; parseLine also trims, but keep
		// the raw line intact for the comment-prefix check.
		line = strings.TrimSuffix(line, "\r")

		key, phones, variant, ok := parseLine(line)
		if !ok {
			continue
		}

		pe := pending[key]
		if pe == nil {
			pe = &pendingEntry{}
			pending[key] = pe
		}
		if variant == 0 {
			// Last unnumbered entry wins deterministically (CMUdict has at most
			// one unnumbered entry per word, so this is normally set once).
			pe.canonical = phones
			pe.hasCanonical = true
		} else {
			pe.alternates = append(pe.alternates, variantPhones{variant: variant, phones: phones})
		}
	}

	entries := make(map[string]*Entry, len(pending))
	for key, pe := range pending {
		// Sort alternates by ascending variant number for determinism.
		sort.SliceStable(pe.alternates, func(i, j int) bool {
			return pe.alternates[i].variant < pe.alternates[j].variant
		})
		alts := make([][]string, 0, len(pe.alternates))
		for _, vp := range pe.alternates {
			alts = append(alts, vp.phones)
		}
		e := &Entry{Alternates: alts}
		if pe.hasCanonical {
			e.Canonical = pe.canonical
		}
		entries[key] = e
	}

	return &Dict{entries: entries}, nil
}

// Lookup resolves word to its Entry using a lowercased key (ADR-0004, ADR-0006).
//
// The variant policy is strict:
//   - A word is found only if it has an unnumbered (canonical) entry.
//   - A word that has only numbered alternates (e.g. only "ONLY(1)") is NOT
//     found — Lookup returns (Entry{}, false). v0.1 never promotes an alternate
//     to canonical.
//   - When found, the returned Entry.Canonical holds the base phonemes and
//     Entry.Alternates holds the remaining variants (currently unused).
func (d *Dict) Lookup(word string) (Entry, bool) {
	key := strings.ToLower(word)
	e, ok := d.entries[key]
	if !ok {
		return Entry{}, false
	}
	// Only words with an unnumbered canonical pronunciation are resolvable.
	if e.Canonical == nil {
		return Entry{}, false
	}
	// Deep-copy to preserve Dict immutability — caller cannot mutate
	// internal backing arrays through the returned Entry.
	canon := make([]string, len(e.Canonical))
	copy(canon, e.Canonical)
	alts := make([][]string, len(e.Alternates))
	for i, alt := range e.Alternates {
		a := make([]string, len(alt))
		copy(a, alt)
		alts[i] = a
	}
	return Entry{Canonical: canon, Alternates: alts}, true
}
