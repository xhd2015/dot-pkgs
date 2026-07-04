package main

import "testing"

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		pattern string
		name    string
		want    bool
	}{
		{"REQUIREMENT-*.md", "REQUIREMENT-DESIGN-wrk-status-compare.md", true},
		{"REQUIREMENT-*.md", "go-pkgs/REQUIREMENT-DESIGN-wrk-status-compare.md", true},
		{".vscode", ".vscode/settings.json", true},
		{".vscode", "vendor/.vscode/extensions.json", true},
		{".agents", "foo/.agents/config.yaml", true},
		{"*.go", "pkg/main.go", true},
		{"*.md", "README.md", true},
		{"REQUIREMENT-*.md", "go-pkgs/README.md", false},
		{"go-pkgs/*.md", "go-pkgs/foo.md", true},
		{"go-pkgs/*.md", "other/foo.md", false},
		{"**/REQUIREMENT-*.md", "go-pkgs/REQUIREMENT-x.md", true},
	}
	for _, tc := range tests {
		got, err := matchesPattern(tc.pattern, tc.name)
		if err != nil {
			t.Fatalf("matchesPattern(%q, %q): %v", tc.pattern, tc.name, err)
		}
		if got != tc.want {
			t.Fatalf("matchesPattern(%q, %q) = %v, want %v", tc.pattern, tc.name, got, tc.want)
		}
	}
}