// Package phoneme defines language-neutral phoneme and phonological feature
// types for Gofonix.
//
// The package exposes the public identity of a phoneme (a neutral integer ID,
// per ADR-0008) and an opaque phonological FeatureMask. ARPAbet and other
// language-specific encodings never appear in this package; they live under
// internal/ (ADR-0001, Principle 3). IPA strings on Phoneme are a debug
// rendering only and carry no stable contract.
//
// In Slice 1 (Phase 1) the package provides only the public types and the
// FeatureMask machinery. The concrete v0.1-en phoneme inventory and feature
// schema table (ADR-0008) are referenced for constants and metadata only; the
// projection of phonemes into feature masks is Slice 2+ work.
package phoneme
