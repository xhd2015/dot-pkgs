package dupstat

import (
	"strings"
	"testing"
)

func TestTokenizeRaw(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "simple return",
			src:  "return nil",
			want: "return nil",
		},
		{
			name: "function call",
			src:  "fmt.Println(\"hello\")",
			want: "fmt . Println ( \"hello\" )",
		},
		{
			name: "assignment",
			src:  "x := 42",
			want: "x := 42",
		},
		{
			name: "if statement",
			src:  "if err != nil { return err }",
			want: "if err != nil { return err }",
		},
		{
			name: "with comments",
			src:  "x := 1 // init\nx = x + 1",
			want: "x := 1 x = x + 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenizeRaw([]byte(tt.src))
			got := strings.Join(tokens, " ")
			if got != tt.want {
				t.Errorf("got  %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestTokenizeNormalized(t *testing.T) {
	t.Run("renamed identifiers", func(t *testing.T) {
		srcA := []byte("foo := bar + 1")
		srcB := []byte("x := y + 1")
		normA := tokenizeNormalized(srcA)
		normB := tokenizeNormalized(srcB)
		got := strings.Join(normA, " ")
		want := strings.Join(normB, " ")
		if got != want {
			t.Errorf("normalized tokens should be equal:\n  A: %q\n  B: %q", got, want)
		}
	})

	t.Run("same identifier maps to same placeholder", func(t *testing.T) {
		src := []byte("x := x + 1")
		norm := tokenizeNormalized(src)
		got := strings.Join(norm, " ")
		parts := strings.Split(got, " ")
		if len(parts) < 3 || parts[0] != parts[2] {
			t.Errorf("same identifier should get same placeholder: %q", got)
		}
	})

	t.Run("keywords unchanged", func(t *testing.T) {
		src := []byte("if err != nil { return err }")
		norm := tokenizeNormalized(src)
		got := strings.Join(norm, " ")
		if !strings.Contains(got, "if") || !strings.Contains(got, "return") {
			t.Errorf("keywords should be preserved: %q", got)
		}
	})
}

func TestTokenizeMixed(t *testing.T) {
	sigSrc := []byte("(a, b int)")
	bodySrc := []byte("if a > 0 { return a } else { return b }")
	tokens := tokenizeMixed(sigSrc, bodySrc)
	got := strings.Join(tokens, " ")
	if !strings.Contains(got, "a") || !strings.Contains(got, "b") || !strings.Contains(got, "int") {
		t.Errorf("signature tokens should be raw: %q", got)
	}
	if !strings.Contains(got, "if") || !strings.Contains(got, "return") {
		t.Errorf("body tokens should be present: %q", got)
	}
}

func TestNormalizeTokensConsistent(t *testing.T) {
	src := []byte("x := x + 1")
	norm := tokenizeNormalized(src)
	got := strings.Join(norm, " ")
	if !strings.Contains(got, "$1") {
		t.Errorf("should contain placeholder: %q", got)
	}
}
