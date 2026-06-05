package arpabet

// StripStress removes a trailing stress digit (0, 1, or 2) from an ARPAbet
// symbol. E.g. "AE1" -> "AE", "IH0" -> "IH", "B" -> "B". Only CMUdict stress
// digits 0/1/2 are stripped; any other trailing character is preserved. The
// input is returned unchanged if it is empty or carries no stress digit.
func StripStress(sym string) string {
	if len(sym) == 0 {
		return sym
	}
	last := sym[len(sym)-1]
	if last == '0' || last == '1' || last == '2' {
		return sym[:len(sym)-1]
	}
	return sym
}
