package pathfmt

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpand(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	tests := map[string]string{
		"":              "",
		"~":             home,
		"~/.grok":       filepath.Join(home, ".grok"),
		"relative/path": "relative/path",
		"/tmp/path":     "/tmp/path",
		"~other/path":   "~other/path",
	}
	for input, want := range tests {
		if got := Expand(input); got != want {
			t.Errorf("Expand(%q) = %q, want %q", input, got, want)
		}
	}
}
