package dupstat

import (
	"strings"
)

func generateNgrams(tokens []string, k int) []string {
	if len(tokens) < k {
		return nil
	}
	ngrams := make([]string, 0, len(tokens)-k+1)
	for i := 0; i <= len(tokens)-k; i++ {
		ngrams = append(ngrams, strings.Join(tokens[i:i+k], "\x00"))
	}
	return ngrams
}

func ngramSet(tokens []string, k int) (map[string]struct{}, bool) {
	ngrams := generateNgrams(tokens, k)
	if len(ngrams) == 0 {
		return nil, false
	}
	set := make(map[string]struct{}, len(ngrams))
	for _, ng := range ngrams {
		set[ng] = struct{}{}
	}
	return set, true
}

func jaccardSimilarity(setA, setB map[string]struct{}) float64 {
	if len(setA) == 0 && len(setB) == 0 {
		return 0.0
	}
	intersection := 0
	for ng := range setA {
		if _, ok := setB[ng]; ok {
			intersection++
		}
	}
	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0.0
	}
	return float64(intersection) / float64(union)
}

func containmentSimilarity(setA, setB map[string]struct{}) float64 {
	if len(setA) == 0 && len(setB) == 0 {
		return 0.0
	}
	intersection := 0
	for ng := range setA {
		if _, ok := setB[ng]; ok {
			intersection++
		}
	}
	minLen := len(setA)
	if len(setB) < minLen {
		minLen = len(setB)
	}
	if minLen == 0 {
		return 0.0
	}
	return float64(intersection) / float64(minLen)
}

func ngramOverlapRatio(setA, setB map[string]struct{}) float64 {
	return jaccardSimilarity(setA, setB)
}

type NgramScores struct {
	JaccardScore    float64
	ContainmentScore float64
	NgramOverlapScore float64
	Valid           bool
}

func computeNgramScores(tokensA, tokensB []string, k int) NgramScores {
	setA, okA := ngramSet(tokensA, k)
	setB, okB := ngramSet(tokensB, k)
	if !okA || !okB {
		return NgramScores{Valid: false}
	}
	return NgramScores{
		JaccardScore:      jaccardSimilarity(setA, setB),
		ContainmentScore:  containmentSimilarity(setA, setB),
		NgramOverlapScore: ngramOverlapRatio(setA, setB),
		Valid:             true,
	}
}

type AllScores struct {
	Raw    NgramScores
	Norm   NgramScores
	Mixed  NgramScores
}
