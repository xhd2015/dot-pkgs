package dupstat

import (
	"math"
	"testing"
)

func TestWordFrequency(t *testing.T) {
	tokens := []string{"ProcessUser", "process", "user", "run"}
	freq := wordFrequency(tokens)
	if freq["process"] != 2 {
		t.Errorf("expected process:2 (from ProcessUser split + literal process), got %d", freq["process"])
	}
	if freq["user"] != 2 {
		t.Errorf("expected user:2 (from ProcessUser split + literal user), got %d", freq["user"])
	}
	if freq["run"] != 1 {
		t.Errorf("expected run:1, got %d", freq["run"])
	}
}

func TestSplitIdentifier(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"ProcessUser", []string{"process", "user"}},
		{"get_user_id", []string{"get", "user", "id"}},
		{"get.user.id", []string{"get", "user", "id"}},
		{"get-user-id", []string{"get", "user", "id"}},
		{"getUserID", []string{"get", "user", "id"}},
		{"HTTPServer", []string{"http", "server"}},
		{"findById", []string{"find", "by", "id"}},
		{"exec", []string{"exec"}},
		{"XMLParser", []string{"xml", "parser"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := splitIdentifier(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("got %v, want %v", got, tt.want)
					return
				}
			}
		})
	}
}

func TestSplitIdentifierNonIdentifier(t *testing.T) {
	nonIDs := []string{"42", "+", "(", "=", ":=", "\"hello\""}
	for _, s := range nonIDs {
		t.Run(s, func(t *testing.T) {
			got := splitIdentifier(s)
			if got != nil {
				t.Errorf("expected nil for %q, got %v", s, got)
			}
		})
	}
}

func TestWordFrequencySplitting(t *testing.T) {
	tokens := []string{"findUser", "=", "saveUser", "(", "userID", ")", "{", "return", "findUser"}
	freq := wordFrequency(tokens)
	if freq["find"] != 2 {
		t.Errorf("expected find:2, got %d", freq["find"])
	}
	if freq["user"] != 4 {
		t.Errorf("expected user:4 (from findUser, saveUser, userID x2), got %d", freq["user"])
	}
	if freq["id"] != 1 {
		t.Errorf("expected id:1 (from userID), got %d", freq["id"])
	}
	if freq["return"] != 1 {
		t.Errorf("expected return:1, got %d", freq["return"])
	}
	if freq["="] != 1 {
		t.Errorf("expected =:1, got %d", freq["="])
	}
}

func TestWeightedJaccard(t *testing.T) {
	freqA := map[string]int{"a": 2, "b": 1}
	freqB := map[string]int{"a": 1, "b": 2, "c": 1}
	score := weightedJaccard(freqA, freqB)
	if score <= 0 || score >= 1.0 {
		t.Errorf("expected score between 0 and 1, got %f", score)
	}
}

func TestWeightedJaccardIdentical(t *testing.T) {
	freqA := map[string]int{"a": 2, "b": 1}
	freqB := map[string]int{"a": 2, "b": 1}
	score := weightedJaccard(freqA, freqB)
	if score != 1.0 {
		t.Errorf("identical freq should give 1.0, got %f", score)
	}
}

func TestWeightedJaccardDisjoint(t *testing.T) {
	freqA := map[string]int{"a": 2}
	freqB := map[string]int{"b": 1}
	score := weightedJaccard(freqA, freqB)
	if score != 0.0 {
		t.Errorf("disjoint freq should give 0.0, got %f", score)
	}
}

func TestWeightedContainment(t *testing.T) {
	freqA := map[string]int{"a": 1, "b": 1}
	freqB := map[string]int{"a": 1, "b": 1, "c": 5}
	score := weightedContainment(freqA, freqB)
	if score != 1.0 {
		t.Errorf("all tokens of A in B should give 1.0, got %f", score)
	}
}

func TestWeightedContainmentPartial(t *testing.T) {
	freqA := map[string]int{"a": 5, "b": 5}
	freqB := map[string]int{"a": 1, "b": 1}
	score := weightedContainment(freqA, freqB)
	if score <= 0 || score >= 1.0 {
		t.Errorf("partial containment should be between 0 and 1, got %f", score)
	}
}

func TestWordOverlap(t *testing.T) {
	freqA := map[string]int{"a": 5, "b": 3, "c": 1}
	freqB := map[string]int{"a": 1, "b": 1, "d": 1}
	score := wordOverlap(freqA, freqB)
	if score <= 0 || score >= 1.0 {
		t.Errorf("partial overlap should be between 0 and 1, got %f", score)
	}
}

func TestWordOverlapIdentical(t *testing.T) {
	freqA := map[string]int{"a": 5, "b": 3}
	freqB := map[string]int{"a": 1, "b": 1}
	score := wordOverlap(freqA, freqB)
	if score != 1.0 {
		t.Errorf("same keys should give 1.0, got %f", score)
	}
}

func TestFlatWords(t *testing.T) {
	tokens := []string{"ProcessUser", "for", "42", "(", "findById"}
	words := flatWords(tokens)
	expected := []string{"process", "user", "for", "42", "(", "find", "by", "id"}
	if len(words) != len(expected) {
		t.Errorf("got %v, want %v", words, expected)
		return
	}
	for i := range words {
		if words[i] != expected[i] {
			t.Errorf("got %v, want %v", words[i], expected[i])
		}
	}
}

func TestComputeWordStats(t *testing.T) {
	funcs := []FunctionTokens{
		{Raw: []string{"return", "err", "ProcessUser"}},
		{Raw: []string{"return", "result", "ProcessUser"}},
		{Raw: []string{"return", "nil"}},
	}
	stats := ComputeWordStats(funcs)
	if stats.TotalDocs != 3 {
		t.Errorf("TotalDocs should be 3, got %d", stats.TotalDocs)
	}
	if stats.DocFreq["return"] != 3 {
		t.Errorf("'return' should appear in 3 docs, got %d", stats.DocFreq["return"])
	}
	if stats.DocFreq["process"] != 2 {
		t.Errorf("'process' (from ProcessUser) should appear in 2 docs, got %d", stats.DocFreq["process"])
	}
	if stats.DocFreq["user"] != 2 {
		t.Errorf("'user' (from ProcessUser) should appear in 2 docs, got %d", stats.DocFreq["user"])
	}
	if stats.DocFreq["err"] != 1 {
		t.Errorf("'err' should appear in 1 doc, got %d", stats.DocFreq["err"])
	}
	if stats.DocFreq["nil"] != 1 {
		t.Errorf("'nil' should appear in 1 doc, got %d", stats.DocFreq["nil"])
	}
}

func TestWordTFIDF(t *testing.T) {
	stats := &WordStats{
		TotalDocs: 3,
		DocFreq: map[string]int{
			"return":  3,
			"process": 1,
			"user":    1,
			"err":     1,
		},
	}
	tokens := []string{"return", "ProcessUser", "err"}
	freq := wordTFIDF(tokens, stats)

	expectReturn := keywordWeight * math.Log(1.0+3.0/4.0)
	expectProcess := identifierWeight * math.Log(1.0+3.0/2.0)
	expectUser := identifierWeight * math.Log(1.0+3.0/2.0)
	expectErr := identifierWeight * math.Log(1.0+3.0/2.0)

	if math.Abs(freq["return"]-expectReturn) > 0.001 {
		t.Errorf("'return' weight: got %f, want %f", freq["return"], expectReturn)
	}
	if math.Abs(freq["process"]-expectProcess) > 0.001 {
		t.Errorf("'process' weight: got %f, want %f", freq["process"], expectProcess)
	}
	if math.Abs(freq["user"]-expectUser) > 0.001 {
		t.Errorf("'user' weight: got %f, want %f", freq["user"], expectUser)
	}
	if math.Abs(freq["err"]-expectErr) > 0.001 {
		t.Errorf("'err' weight: got %f, want %f", freq["err"], expectErr)
	}

	if freq["return"] >= freq["process"] {
		t.Errorf("common word 'return' should have lower weight than rare word 'process': %f >= %f",
			freq["return"], freq["process"])
	}
}

func TestWeightedWordFrequencyKeywordsAndOperators(t *testing.T) {
	tokens := []string{"return", "for", "if", ":=", "==", "(", ")"}
	freq := weightedWordFrequency(tokens)
	for _, tok := range tokens {
		if freq[tok] != keywordWeight && freq[tok] != operatorWeight {
			t.Errorf("expected %s to have weight %f (keyword) or %f (operator), got %f",
				tok, keywordWeight, operatorWeight, freq[tok])
		}
	}
	if freq["return"] != keywordWeight {
		t.Errorf("'return' keyword: got %f, want %f", freq["return"], keywordWeight)
	}
	if freq[":="] != operatorWeight {
		t.Errorf("':=' operator: got %f, want %f", freq[":="], operatorWeight)
	}
}

func TestWeightedWordFrequencyLiterals(t *testing.T) {
	tokens := []string{"\"hello\"", "42", "`raw`", "'c'"}
	freq := weightedWordFrequency(tokens)
	for _, tok := range tokens {
		if freq[tok] != literalWeight {
			t.Errorf("expected literal %s to have weight %f, got %f", tok, literalWeight, freq[tok])
		}
	}
}

func TestWeightedWordFrequencyIdentifiers(t *testing.T) {
	tokens := []string{"ProcessUser", "findById", "err"}
	freq := weightedWordFrequency(tokens)
	if freq["process"] != identifierWeight {
		t.Errorf("'process' (from ProcessUser): got %f, want %f", freq["process"], identifierWeight)
	}
	if freq["user"] != identifierWeight {
		t.Errorf("'user' (from ProcessUser): got %f, want %f", freq["user"], identifierWeight)
	}
	if freq["find"] != identifierWeight {
		t.Errorf("'find' (from findById): got %f, want %f", freq["find"], identifierWeight)
	}
	if freq["by"] != identifierWeight {
		t.Errorf("'by' (from findById): got %f, want %f", freq["by"], identifierWeight)
	}
	if freq["id"] != identifierWeight {
		t.Errorf("'id' (from findById): got %f, want %f", freq["id"], identifierWeight)
	}
	if freq["err"] != identifierWeight {
		t.Errorf("'err': got %f, want %f", freq["err"], identifierWeight)
	}
}

func TestWeightedWordFrequencyFiltersSingleLetterSplitArtifacts(t *testing.T) {
	tokens := []string{"freqA", "freqB"}
	freq := weightedWordFrequency(tokens)

	if _, ok := freq["freq"]; !ok {
		t.Error("'freq' (from freqA) should be present")
	}
	if _, ok := freq["a"]; ok {
		t.Error("'a' from split freqA should be filtered")
	}
	if _, ok := freq["b"]; ok {
		t.Error("'b' from split freqB should be filtered")
	}

	if freq["freq"] != 2*identifierWeight {
		t.Errorf("'freq' appears twice (from freqA+freqB): got %f, want %f", freq["freq"], 2*identifierWeight)
	}
}

func TestWeightedWordFrequencyKeepsSingleLetterIdentifiers(t *testing.T) {
	tokens := []string{"a", "b", "k", "i"}
	freq := weightedWordFrequency(tokens)
	for _, tok := range tokens {
		if freq[tok] != identifierWeight {
			t.Errorf("single-letter identifier %s should be kept: got %f, want %f", tok, freq[tok], identifierWeight)
		}
	}
}

func TestIsNoiseWord(t *testing.T) {
	if !isNoiseWord("return") {
		t.Errorf("'return' (keyword) should be noise")
	}
	if !isNoiseWord(":=") {
		t.Errorf("':=' (operator) should be noise")
	}
	if !isNoiseWord("(") {
		t.Errorf("'(' (operator) should be noise")
	}
	if isNoiseWord("process") {
		t.Errorf("'process' (identifier word) should NOT be noise")
	}
	if isNoiseWord("\"hello\"") {
		t.Errorf("'\"hello\"' (literal) should NOT be noise")
	}
}

func TestWeightedJaccardFloatIdentical(t *testing.T) {
	freqA := map[string]float64{"a": 2.0, "b": 1.0}
	freqB := map[string]float64{"a": 2.0, "b": 1.0}
	score := weightedJaccardFloat(freqA, freqB)
	if score != 1.0 {
		t.Errorf("identical freq should give 1.0, got %f", score)
	}
}

func TestWeightedJaccardFloatDisjoint(t *testing.T) {
	freqA := map[string]float64{"a": 2.0}
	freqB := map[string]float64{"b": 1.0}
	score := weightedJaccardFloat(freqA, freqB)
	if score != 0.0 {
		t.Errorf("disjoint freq should give 0.0, got %f", score)
	}
}

func TestWeightedContainmentFloatFull(t *testing.T) {
	freqA := map[string]float64{"a": 1.0, "b": 1.0}
	freqB := map[string]float64{"a": 2.0, "b": 2.0, "c": 5.0}
	score := weightedContainmentFloat(freqA, freqB)
	if score != 1.0 {
		t.Errorf("all tokens of A in B should give 1.0, got %f", score)
	}
}

func TestWeightedContainmentFloatPartial(t *testing.T) {
	freqA := map[string]float64{"a": 5.0, "b": 5.0}
	freqB := map[string]float64{"a": 1.0, "b": 1.0}
	score := weightedContainmentFloat(freqA, freqB)
	if score <= 0 || score >= 1.0 {
		t.Errorf("partial containment should be between 0 and 1, got %f", score)
	}
}

func TestWordOverlapFloatFiltersKeywords(t *testing.T) {
	freqA := map[string]float64{"process": 1.0, "for": 3.0, "if": 2.0, "return": 1.0}
	freqB := map[string]float64{"process": 2.0, "save": 1.0, "for": 1.0, "return": 2.0}
	score := wordOverlapFloat(freqA, freqB)
	expected := 0.5
	if math.Abs(score-expected) > 0.001 {
		t.Errorf("after filtering keywords: intersect={process}, union={process,save} => 1/2=0.5, got %f", score)
	}
}

func TestWordOverlapFloatAllKeywords(t *testing.T) {
	freqA := map[string]float64{"for": 3.0, "if": 2.0, "return": 1.0}
	freqB := map[string]float64{"for": 1.0, "if": 1.0, "return": 2.0}
	score := wordOverlapFloat(freqA, freqB)
	if score != 0.0 {
		t.Errorf("when all words are keywords, filtered maps are empty, overlap should be 0, got %f", score)
	}
}

func TestWeightedContainmentFloatAsymmetric(t *testing.T) {
	freqSmall := map[string]float64{"a": 2.0, "return": 0.5}
	freqLarge := map[string]float64{"a": 1.0, "return": 0.5, "process": 5.0, "user": 3.0, "validate": 2.0}

	cSmallToLarge := weightedContainmentFloat(freqSmall, freqLarge)
	cLargeToSmall := weightedContainmentFloat(freqLarge, freqSmall)

	if cSmallToLarge <= cLargeToSmall {
		t.Errorf("small→large containment should be high (>0.5), got %f", cSmallToLarge)
	}
	if cLargeToSmall >= 0.3 {
		t.Errorf("large→small containment should be low (<0.3), got %f", cLargeToSmall)
	}

	bidirectional := cSmallToLarge
	if cLargeToSmall < bidirectional {
		bidirectional = cLargeToSmall
	}
	if bidirectional >= 0.3 {
		t.Errorf("bidirectional containment should be <0.3 for asymmetric pair, got %f", bidirectional)
	}
}
