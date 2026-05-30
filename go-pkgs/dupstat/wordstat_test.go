package dupstat

import (
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
