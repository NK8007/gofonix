package dict

type Dictionary map[string]string

func (d Dictionary) Lookup(word string) (string, bool) {
	phonemes, ok := d[word]
	return phonemes, ok
}
