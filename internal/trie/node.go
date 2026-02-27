package trie

// Node represents a single character/prefix in the trie
type Node struct {
	Children    map[rune]*Node
	Value       string
	IsEndOfWord bool
}
