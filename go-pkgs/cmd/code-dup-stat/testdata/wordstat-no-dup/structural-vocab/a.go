package structuralvocab

func NormalizeWords(tokens []string) []string {
	seen := make(map[string]int)
	result := make([]string, len(tokens))
	for i, tok := range tokens {
		if _, ok := seen[tok]; !ok {
			seen[tok] = len(seen) + 1
		}
		result[i] = string(rune('0' + seen[tok]))
	}
	return result
}
