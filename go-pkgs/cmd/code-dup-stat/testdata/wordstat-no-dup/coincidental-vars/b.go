package coincidentalvars

func WeightedSum(a, b map[string]float64) float64 {
	var result float64
	for k := range a {
		va := a[k]
		vb := b[k]
		if va < vb {
			result += va
		} else {
			result += vb
		}
	}
	return result
}
