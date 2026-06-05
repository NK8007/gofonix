package phoneme

// Feature is an index into a phonological feature schema (ADR-0008). Each
// Feature value corresponds to a single bit position within a FeatureMask.
//
// In v0.1-en the valid range is 0-22 (23 features); bits 23-127 are reserved.
// Slice 1 defines the type only; the concrete schema table is Slice 2+.
type Feature int
