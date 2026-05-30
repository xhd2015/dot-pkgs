package dupstat

import (
	"math"
	"testing"
)

func TestGenerateNgrams(t *testing.T) {
	tokens := []string{"if", "err", "!=", "nil", "{", "return", "err", "}"}
	ngrams := generateNgrams(tokens, 3)
	if len(ngrams) != 6 {
		t.Errorf("expected 6 ngrams for k=3 with 8 tokens, got %d", len(ngrams))
	}
}

func TestGenerateNgramsShort(t *testing.T) {
	tokens := []string{"a", "b"}
	ngrams := generateNgrams(tokens, 3)
	if len(ngrams) != 0 {
		t.Errorf("expected 0 ngrams when tokens < k, got %d", len(ngrams))
	}
}

func TestGenerateNgramsExact(t *testing.T) {
	tokens := []string{"a", "b", "c"}
	ngrams := generateNgrams(tokens, 3)
	if len(ngrams) != 1 {
		t.Errorf("expected 1 ngram, got %d", len(ngrams))
	}
}

func TestJaccardIdentical(t *testing.T) {
	tokens := []string{"a", "b", "c", "d", "e"}
	setA, _ := ngramSet(tokens, 3)
	setB, _ := ngramSet(tokens, 3)
	score := jaccardSimilarity(setA, setB)
	if score != 1.0 {
		t.Errorf("identical sets should have jaccard=1.0, got %f", score)
	}
}

func TestJaccardDisjoint(t *testing.T) {
	setA, _ := ngramSet([]string{"a", "b", "c", "d", "e"}, 3)
	setB, _ := ngramSet([]string{"f", "g", "h", "i", "j"}, 3)
	score := jaccardSimilarity(setA, setB)
	if score != 0.0 {
		t.Errorf("disjoint sets should have jaccard=0.0, got %f", score)
	}
}

func TestJaccardEmpty(t *testing.T) {
	setA := map[string]struct{}{}
	setB := map[string]struct{}{}
	score := jaccardSimilarity(setA, setB)
	if score != 0.0 {
		t.Errorf("two empty sets should have jaccard=0.0, got %f", score)
	}
}

func TestContainmentFull(t *testing.T) {
	tokensA := []string{"a", "b", "c", "d", "e"}
	tokensB := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}
	setA, _ := ngramSet(tokensA, 2)
	setB, _ := ngramSet(tokensB, 2)
	score := containmentSimilarity(setA, setB)
	if score != 1.0 {
		t.Errorf("when all ngrams of smaller are in larger, containment should be 1.0, got %f", score)
	}
}

func TestContainmentPartial(t *testing.T) {
	tokensA := []string{"a", "b", "c", "d", "e"}
	tokensB := []string{"a", "b", "f", "g", "h"}
	setA, _ := ngramSet(tokensA, 2)
	setB, _ := ngramSet(tokensB, 2)
	score := containmentSimilarity(setA, setB)
	if score <= 0 || score >= 1.0 {
		t.Errorf("partial containment should be between 0 and 1, got %f", score)
	}
}

func TestNgramSetShort(t *testing.T) {
	tokens := []string{"a", "b"}
	_, ok := ngramSet(tokens, 5)
	if ok {
		t.Errorf("should return ok=false for tokens shorter than k")
	}
}

func TestComputeNgramScoresIdentical(t *testing.T) {
	tokensA := []string{"if", "err", "!=", "nil", "{", "return", "err", "}"}
	tokensB := []string{"if", "err", "!=", "nil", "{", "return", "err", "}"}
	scores := computeNgramScores(tokensA, tokensB, 3)
	if scores.JaccardScore != 1.0 {
		t.Errorf("jaccard should be 1.0 for identical token sequences, got %f", scores.JaccardScore)
	}
	if scores.ContainmentScore != 1.0 {
		t.Errorf("containment should be 1.0 for identical token sequences, got %f", scores.ContainmentScore)
	}
}

func TestComputeNgramScoresDifferent(t *testing.T) {
	tokensA := []string{"a", "b", "c", "d", "e"}
	tokensB := []string{"x", "y", "z", "w", "v"}
	scores := computeNgramScores(tokensA, tokensB, 3)
	if math.Abs(scores.JaccardScore-0.0) > 0.01 {
		t.Errorf("jaccard should be ~0 for different sequences, got %f", scores.JaccardScore)
	}
}
