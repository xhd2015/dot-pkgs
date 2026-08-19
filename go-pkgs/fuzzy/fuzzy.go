package fuzzy

import (
	"strings"
	"unicode"
)

const (
	scoreMatch        = 16
	scoreGapStart     = -3
	scoreGapExtension = -1

	bonusBoundary            = scoreMatch / 2 // 8
	bonusNonWord             = scoreMatch / 2 // 8
	bonusCamel123            = bonusBoundary + scoreGapExtension
	bonusConsecutive         = -(scoreGapStart + scoreGapExtension) // 4
	bonusFirstCharMultiplier = 2
	bonusBoundaryWhite       = bonusBoundary + 2 // 10
	bonusBoundaryDelimiter   = bonusBoundary + 1 // 9
	bonusPathDelimiter       = bonusBoundaryDelimiter + 1
)

type charClass int

const (
	charWhite charClass = iota
	charNonWord
	charDelimiter
	charLower
	charUpper
	charLetter
	charNumber
)

// Tokens splits query on unicode whitespace and drops empty fields.
func Tokens(query string) []string {
	return strings.Fields(query)
}

// Match scores one literal token (spaces in query are required characters).
func Match(haystack, query string, opts ...Option) Result {
	cfg := applyOptions(opts)
	if query == "" {
		return emptyResult(haystack)
	}
	ok, score, idxs := matchOne(haystack, query, cfg)
	if !ok {
		return Result{OK: false}
	}
	return Result{OK: true, Score: score, Spans: spansFromIndexes([]rune(haystack), idxs)}
}

// MatchAll requires every token to match. Score is the sum of token scores;
// matched rune indexes are the union of per-token matches.
func MatchAll(haystack string, tokens []string, opts ...Option) Result {
	if len(tokens) == 0 {
		return emptyResult(haystack)
	}
	cfg := applyOptions(opts)
	hRunes := []rune(haystack)
	union := make([]bool, len(hRunes))
	score := 0
	for _, tok := range tokens {
		if tok == "" {
			continue
		}
		ok, tokScore, idxs := matchOne(haystack, tok, cfg)
		if !ok {
			return Result{OK: false}
		}
		score += tokScore
		for _, i := range idxs {
			if i >= 0 && i < len(union) {
				union[i] = true
			}
		}
	}
	return Result{OK: true, Score: score, Spans: spansFromMatched(hRunes, union)}
}

func emptyResult(haystack string) Result {
	return Result{
		OK:    true,
		Score: 0,
		Spans: []Span{{Text: haystack, Matched: false}},
	}
}

func matchOne(haystack, query string, cfg config) (bool, int, []int) {
	hRunes := []rune(haystack)
	qRunes := []rune(query)
	if len(qRunes) == 0 {
		return true, 0, nil
	}
	if len(qRunes) > len(hRunes) {
		return false, 0, nil
	}

	hFold := foldRunes(hRunes, cfg.caseSensitive)
	qFold := foldRunes(qRunes, cfg.caseSensitive)

	// Forward scan: first subsequence that consumes every query rune.
	pidx := 0
	sidx := -1
	eidx := -1
	for i, r := range hFold {
		if r != qFold[pidx] {
			continue
		}
		if sidx < 0 {
			sidx = i
		}
		pidx++
		if pidx == len(qFold) {
			eidx = i + 1
			break
		}
	}
	if sidx < 0 || eidx < 0 {
		return false, 0, nil
	}

	// Backward tighten to the shortest window that still contains the pattern.
	pidx = len(qFold) - 1
	for i := eidx - 1; i >= sidx; i-- {
		if hFold[i] != qFold[pidx] {
			continue
		}
		pidx--
		if pidx < 0 {
			sidx = i
			break
		}
	}

	score, pos := calculateScore(hRunes, hFold, qFold, sidx, eidx, cfg)
	return true, score, pos
}

func foldRunes(rs []rune, caseSensitive bool) []rune {
	if caseSensitive {
		out := make([]rune, len(rs))
		copy(out, rs)
		return out
	}
	out := make([]rune, len(rs))
	for i, r := range rs {
		out[i] = unicode.ToLower(r)
	}
	return out
}

func calculateScore(hOrig, hFold, qFold []rune, sidx, eidx int, cfg config) (int, []int) {
	pidx, score, consecutive, firstBonus := 0, 0, 0, 0
	inGap := false
	pos := make([]int, 0, len(qFold))

	prevClass := charWhite
	if cfg.pathScheme {
		prevClass = charDelimiter
	}
	if sidx > 0 {
		prevClass = classify(hOrig[sidx-1], cfg)
	}

	for idx := sidx; idx < eidx; idx++ {
		class := classify(hOrig[idx], cfg)
		if pidx < len(qFold) && hFold[idx] == qFold[pidx] {
			pos = append(pos, idx)
			score += scoreMatch
			bonus := bonusFor(prevClass, class, cfg)
			if consecutive == 0 {
				firstBonus = bonus
			} else {
				if bonus >= bonusBoundary && bonus > firstBonus {
					firstBonus = bonus
				}
				if bonus < firstBonus {
					bonus = firstBonus
				}
				if bonus < bonusConsecutive {
					bonus = bonusConsecutive
				}
			}
			if pidx == 0 {
				score += bonus * bonusFirstCharMultiplier
			} else {
				score += bonus
			}
			inGap = false
			consecutive++
			pidx++
		} else {
			if inGap {
				score += scoreGapExtension
			} else {
				score += scoreGapStart
			}
			inGap = true
			consecutive = 0
			firstBonus = 0
		}
		prevClass = class
	}
	return score, pos
}

func classify(r rune, cfg config) charClass {
	if unicode.IsSpace(r) {
		return charWhite
	}
	if isPathSep(r) && cfg.pathScheme {
		return charDelimiter
	}
	if _, ok := cfg.separators[r]; ok {
		return charDelimiter
	}
	if unicode.IsLower(r) {
		return charLower
	}
	if unicode.IsUpper(r) {
		return charUpper
	}
	if unicode.IsNumber(r) {
		return charNumber
	}
	if unicode.IsLetter(r) {
		return charLetter
	}
	return charNonWord
}

func bonusFor(prev, class charClass, cfg config) int {
	if class > charDelimiter {
		switch prev {
		case charWhite:
			if cfg.pathScheme {
				return bonusBoundary
			}
			return bonusBoundaryWhite
		case charDelimiter:
			if cfg.pathScheme {
				return bonusPathDelimiter
			}
			return bonusBoundaryDelimiter
		case charNonWord:
			return bonusBoundary
		}
	}
	if prev == charLower && class == charUpper || prev != charNumber && class == charNumber {
		return bonusCamel123
	}
	switch class {
	case charNonWord, charDelimiter:
		return bonusNonWord
	case charWhite:
		if cfg.pathScheme {
			return bonusBoundary
		}
		return bonusBoundaryWhite
	}
	return 0
}

func spansFromIndexes(runes []rune, idxs []int) []Span {
	matched := make([]bool, len(runes))
	for _, i := range idxs {
		if i >= 0 && i < len(matched) {
			matched[i] = true
		}
	}
	return spansFromMatched(runes, matched)
}

func spansFromMatched(runes []rune, matched []bool) []Span {
	if len(runes) == 0 {
		return []Span{{Text: "", Matched: false}}
	}
	var spans []Span
	var b strings.Builder
	cur := matched[0]
	for i, r := range runes {
		if matched[i] != cur {
			spans = append(spans, Span{Text: b.String(), Matched: cur})
			b.Reset()
			cur = matched[i]
		}
		b.WriteRune(r)
	}
	spans = append(spans, Span{Text: b.String(), Matched: cur})
	return spans
}
