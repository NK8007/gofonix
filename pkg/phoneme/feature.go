package phoneme

// Feature is an index into a phonological feature schema (ADR-0008). Each
// Feature value corresponds to a single bit position within a FeatureMask.
//
// In v0.1-en the valid range is 0-22 (23 features); bits 23-127 are reserved.
// The concrete schema table is the v0.1-en assignment defined in ADR-0008.
type Feature int
