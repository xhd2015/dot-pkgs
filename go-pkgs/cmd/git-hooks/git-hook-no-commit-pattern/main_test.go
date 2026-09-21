package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestRunAutoUnstageMatchedOnFreshRepo(t *testing.T) {
	// Zero commits: no HEAD for git restore --staged; the unstage must
	// fall back to git rm --cached.
	repo := initPatternRepo(t)
	t.Chdir(repo)

	writePatternFile(t, filepath.Join(repo, "keep.txt"), "hello\n")
	writePatternFile(t, filepath.Join(repo, "app.min.js"), "var x=1;\n")
	mustPatternRun(t, repo, "git", "add", "keep.txt", "app.min.js")

	var out bytes.Buffer
	err := runWithOutput([]string{"--auto-unstage", "*.min.js"}, &out)
	if err != nil {
		t.Fatalf("expected no error with --auto-unstage on repo without commits, got %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "auto-unstaged 1 file(s) (kept on disk)") {
		t.Fatalf("expected auto-unstaged note in output, got:\n%s", out.String())
	}

	staged := patternStagedFiles(t)
	if len(staged) != 1 || staged[0] != "keep.txt" {
		t.Fatalf("expected only keep.txt staged, got %v", staged)
	}
	if _, err := os.Stat(filepath.Join(repo, "app.min.js")); err != nil {
		t.Fatalf("app.min.js must remain on disk: %v", err)
	}
}

func TestRunAutoUnstageMatchedWithCommit(t *testing.T) {
	repo := initPatternRepo(t)
	t.Chdir(repo)

	writePatternFile(t, filepath.Join(repo, "keep.txt"), "hello\n")
	writePatternFile(t, filepath.Join(repo, "app.min.js"), "var x=1;\n")
	mustPatternRun(t, repo, "git", "add", "keep.txt", "app.min.js")
	mustPatternRun(t, repo, "git", "commit", "-m", "init")

	writePatternFile(t, filepath.Join(repo, "app.min.js"), "var x=2;\n")
	mustPatternRun(t, repo, "git", "add", "app.min.js")

	var out bytes.Buffer
	err := runWithOutput([]string{"--auto-unstage", "*.min.js"}, &out)
	if err != nil {
		t.Fatalf("expected no error with --auto-unstage, got %v\n%s", err, out.String())
	}

	staged := patternStagedFiles(t)
	if len(staged) != 0 {
		t.Fatalf("expected nothing staged after unstage, got %v", staged)
	}
	cmd := exec.Command("git", "ls-files", "--", "app.min.js")
	output, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(output)) == "" {
		t.Errorf("app.min.js must remain tracked (restore --staged, not rm --cached)")
	}
}

func initPatternRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	mustPatternRun(t, repo, "git", "init")
	mustPatternRun(t, repo, "git", "config", "user.email", "test@example.com")
	mustPatternRun(t, repo, "git", "config", "user.name", "Test User")
	return repo
}

func writePatternFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func mustPatternRun(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, output)
	}
}

func patternStagedFiles(t *testing.T) []string {
	t.Helper()
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git diff --cached --name-only: %v", err)
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	return files
}