package phoneme

// ToggleFeature applies bitwise operations to phonetic feature masks
func ToggleFeature(mask, feature uint16) uint16 {
	return mask ^ feature
}
