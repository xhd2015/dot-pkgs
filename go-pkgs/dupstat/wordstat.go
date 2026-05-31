package dupstat

import (
	"math"
	"strings"
	"unicode"
)

func wordFrequency(tokens []string) map[string]int {
	freq := make(map[string]int)
	for _, t := range tokens {
		words := splitIdentifier(t)
		if len(words) == 0 {
			freq[t]++
		} else {
			for _, w := range words {
				freq[w]++
			}
		}
	}
	return freq
}

const (
	keywordWeight    = 0.05
	operatorWeight   = 0.05
	literalWeight    = 2.0
	identifierWeight = 1.5
)

func weightedWordFrequency(tokens []string) map[string]float64 {
	freq := make(map[string]float64)
	for _, t := range tokens {
		if goKeywords[t] {
			freq[t] += keywordWeight
			continue
		}
		if isOperator(t) {
			freq[t] += operatorWeight
			continue
		}
		if isLiteral(t) {
			freq[t] += literalWeight
			continue
		}
		words := splitIdentifier(t)
		if len(words) <= 1 {
			for _, w := range words {
				freq[w] += identifierWeight
			}
			if len(words) == 0 {
				freq[t] += identifierWeight
			}
			continue
		}
		for _, w := range words {
			if len(w) > 1 {
				freq[w] += identifierWeight
			}
		}
	}
	return freq
}

func isNoiseWord(s string) bool {
	return goKeywords[s] || isOperator(s)
}

func splitIdentifier(s string) []string {
	if len(s) == 0 {
		return nil
	}
	first := rune(s[0])
	if !unicode.IsLetter(first) {
		return nil
	}

	var words []string
	var buf []rune

	flush := func() {
		if len(buf) > 0 {
			words = append(words, strings.ToLower(string(buf)))
			buf = buf[:0]
		}
	}

	runes := []rune(s)
	for i, r := range runes {
		if r == '_' || r == '.' || r == '-' || unicode.IsDigit(r) {
			flush()
			continue
		}

		if i > 0 {
			prev := runes[i-1]
			// lower -> UPPER: processUser → process, User
			if unicode.IsLower(prev) && unicode.IsUpper(r) {
				flush()
			}
			// UPPER UPPER ... UPPER lower: HTTPServer → HTTP, Server
			if unicode.IsUpper(prev) && unicode.IsLower(r) && len(buf) > 1 {
				last := buf[len(buf)-1]
				if unicode.IsUpper(last) {
					buf = buf[:len(buf)-1]
					flush()
					buf = append(buf, last)
				}
			}
		}

		buf = append(buf, r)
	}

	flush()
	return words
}

func weightedJaccard(freqA, freqB map[string]int) float64 {
	var sumMin, sumMax int
	allKeys := make(map[string]struct{})
	for k := range freqA {
		allKeys[k] = struct{}{}
	}
	for k := range freqB {
		allKeys[k] = struct{}{}
	}
	for k := range allKeys {
		a := freqA[k]
		b := freqB[k]
		if a < b {
			sumMin += a
		} else {
			sumMin += b
		}
		if a > b {
			sumMax += a
		} else {
			sumMax += b
		}
	}
	if sumMax == 0 {
		return 0.0
	}
	return float64(sumMin) / float64(sumMax)
}

func weightedContainment(freqA, freqB map[string]int) float64 {
	var sumMin, sumA int
	for k := range freqA {
		a := freqA[k]
		b := freqB[k]
		if a < b {
			sumMin += a
		} else {
			sumMin += b
		}
		sumA += a
	}
	if sumA == 0 {
		return 0.0
	}
	return float64(sumMin) / float64(sumA)
}

func wordOverlap(freqA, freqB map[string]int) float64 {
	intersection := 0
	union := len(freqB)
	for k := range freqA {
		if _, ok := freqB[k]; ok {
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

type WordStats struct {
	TotalDocs int
	DocFreq   map[string]int
}

func ComputeWordStats(allFuncs []FunctionTokens) *WordStats {
	stats := &WordStats{
		TotalDocs: len(allFuncs),
		DocFreq:   make(map[string]int),
	}
	for _, ft := range allFuncs {
		seen := make(map[string]bool)
		for _, word := range flatWords(ft.Raw) {
			if !seen[word] {
				seen[word] = true
				stats.DocFreq[word]++
			}
		}
	}
	return stats
}

func flatWords(tokens []string) []string {
	var words []string
	for _, t := range tokens {
		sub := splitIdentifier(t)
		if len(sub) == 0 {
			words = append(words, t)
		} else {
			words = append(words, sub...)
		}
	}
	return words
}

func wordTFIDF(tokens []string, stats *WordStats) map[string]float64 {
	tf := weightedWordFrequency(tokens)
	result := make(map[string]float64)
	for w, count := range tf {
		df := stats.DocFreq[w]
		idf := math.Log(1.0 + float64(stats.TotalDocs)/float64(1+df))
		result[w] = count * idf
	}
	return result
}

func weightedJaccardFloat(freqA, freqB map[string]float64) float64 {
	var sumMin, sumMax float64
	allKeys := make(map[string]struct{})
	for k := range freqA {
		allKeys[k] = struct{}{}
	}
	for k := range freqB {
		allKeys[k] = struct{}{}
	}
	for k := range allKeys {
		a := freqA[k]
		b := freqB[k]
		if a < b {
			sumMin += a
		} else {
			sumMin += b
		}
		if a > b {
			sumMax += a
		} else {
			sumMax += b
		}
	}
	if sumMax == 0 {
		return 0.0
	}
	return sumMin / sumMax
}

func weightedContainmentFloat(freqA, freqB map[string]float64) float64 {
	var sumMin, sumA float64
	for k := range freqA {
		a := freqA[k]
		b := freqB[k]
		if a < b {
			sumMin += a
		} else {
			sumMin += b
		}
		sumA += a
	}
	if sumA == 0 {
		return 0.0
	}
	return sumMin / sumA
}

func wordOverlapFloat(freqA, freqB map[string]float64) float64 {
	filteredA := make(map[string]bool)
	filteredB := make(map[string]bool)
	for k := range freqA {
		if !isNoiseWord(k) {
			filteredA[k] = true
		}
	}
	for k := range freqB {
		if !isNoiseWord(k) {
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
