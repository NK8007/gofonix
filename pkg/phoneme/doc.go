// Package phoneme defines language-neutral phoneme and phonological feature
// types for Gofonix.
//
// The package exposes the public identity of a phoneme (a neutral integer ID,
// per ADR-0008) and an opaque phonological FeatureMask. ARPAbet and other
// language-specific encodings never appear in this package; they live under
// internal/ (ADR-0001, ADR-0008: ARPAbet stays internal). IPA strings on
// Phoneme are a debug rendering only and carry no stable contract.
//
// The package provides the public types and the FeatureMask machinery. The
// concrete v0.1-en phoneme inventory and feature schema table are defined in
// ADR-0008; phonemes are projected into feature masks by the internal English
// ARPAbet bridge.
package phoneme
