package wrappervslarge

func ComputeOverlap(a, b map[string]float64) float64 {
	filteredA := make(map[string]bool)
	filteredB := make(map[string]bool)
	for k := range a {
		if k[0] != '_' {
			filteredA[k] = true
		}
	}
	for k := range b {
		if k[0] != '_' {
			filteredB[k] = true
		}
	}
	intersection := 0
	union := len(filteredB)
	for k := range filteredA {
		if filteredB[k] {
			intersection++
		} else {
			union++
		}
	}
	if union == 0 {
		return 0.0
	}
	return float64(intersection) / float64(union)
}
