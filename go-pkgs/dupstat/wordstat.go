package dupstat

import (
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
