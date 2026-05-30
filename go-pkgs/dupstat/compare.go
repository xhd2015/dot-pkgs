package dupstat

const (
	AlgoNgram    = "ngram"
	AlgoWordstat = "wordstat"
)

type Group struct {
	Kind  string
	Pairs []FuncPair
}

type FuncPair struct {
	FuncA             *Function
	FuncB             *Function
	RawJaccard        float64
	RawContainment    float64
	RawNgramOverlap   float64
	NormJaccard       float64
	NormContainment   float64
	NormNgramOverlap  float64
	MixedJaccard      float64
	MixedContainment  float64
	MixedNgramOverlap float64
	WordJaccard       float64
	WordContainment   float64
	WordOverlap       float64
}

func CompareFunctions(allFuncs []FunctionTokens, k int, threshold float64, algorithm string) []FuncPair {
	var pairs []FuncPair
	for i := 0; i < len(allFuncs); i++ {
		for j := i + 1; j < len(allFuncs); j++ {
			a := allFuncs[i]
			b := allFuncs[j]

			if a.Func.File == b.Func.File {
				continue
			}

			var pair FuncPair
			pair.FuncA = a.Func
			pair.FuncB = b.Func

			if algorithm == AlgoWordstat {
				if !computeWordstatPair(&pair, a.Raw, b.Raw) {
					continue
				}
				maxScore := max3(pair.WordJaccard, pair.WordContainment, pair.WordOverlap)
				if maxScore < threshold {
					continue
				}
			} else {
				raw := computeNgramScores(a.Raw, b.Raw, k)
				norm := computeNgramScores(a.Norm, b.Norm, k)
				mixed := computeNgramScores(a.Mixed, b.Mixed, k)

				if !raw.Valid && !norm.Valid && !mixed.Valid {
					continue
				}

				pair.RawJaccard = raw.JaccardScore
				pair.RawContainment = raw.ContainmentScore
				pair.RawNgramOverlap = raw.NgramOverlapScore
				pair.NormJaccard = norm.JaccardScore
				pair.NormContainment = norm.ContainmentScore
				pair.NormNgramOverlap = norm.NgramOverlapScore
				pair.MixedJaccard = mixed.JaccardScore
				pair.MixedContainment = mixed.ContainmentScore
				pair.MixedNgramOverlap = mixed.NgramOverlapScore

				maxScore := 0.0
				for _, s := range []float64{raw.JaccardScore, raw.ContainmentScore, norm.JaccardScore, norm.ContainmentScore, mixed.JaccardScore, mixed.ContainmentScore} {
					if s > maxScore {
						maxScore = s
					}
				}

				if maxScore < threshold {
					continue
				}
			}

			pairs = append(pairs, pair)
		}
	}
	return pairs
}

func computeWordstatPair(pair *FuncPair, tokensA, tokensB []string) bool {
	if len(tokensA) == 0 || len(tokensB) == 0 {
		return false
	}
	freqA := wordFrequency(tokensA)
	freqB := wordFrequency(tokensB)
	if len(freqA) == 0 || len(freqB) == 0 {
		return false
	}
	pair.WordJaccard = weightedJaccard(freqA, freqB)
	pair.WordContainment = weightedContainment(freqA, freqB)
	pair.WordOverlap = wordOverlap(freqA, freqB)
	return true
}

func GroupPairs(pairs []FuncPair) []Group {
	var crossPkgPairs, samePkgPairs []FuncPair
	for _, p := range pairs {
		pkgA := p.FuncA.PkgPath
		pkgB := p.FuncB.PkgPath
		if pkgA != pkgB {
			crossPkgPairs = append(crossPkgPairs, p)
		} else {
			samePkgPairs = append(samePkgPairs, p)
		}
	}

	var groups []Group
	if len(crossPkgPairs) > 0 {
		groups = append(groups, Group{Kind: "cross-package", Pairs: crossPkgPairs})
	}
	if len(samePkgPairs) > 0 {
		groups = append(groups, Group{Kind: "same-package", Pairs: samePkgPairs})
	}
	return groups
}

func max3(a, b, c float64) float64 {
	if a >= b && a >= c {
		return a
	}
	if b >= c {
		return b
	}
	return c
}
