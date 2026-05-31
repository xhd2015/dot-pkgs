package structuralvocab

func FilterNoise(words []string) []string {
	var result []string
	for _, w := range words {
		if len(w) == 0 {
			continue
		}
		if w[0] == '_' || w[0] == '$' {
			continue
		}
		result = append(result, w)
	}
	return result
}
