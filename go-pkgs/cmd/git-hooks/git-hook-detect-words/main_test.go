package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	githook "github.com/xhd2015/dot-pkgs/go-pkgs/git-hook"
)

func TestRunDetectsStagedAddedLines(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	writeFile(t, filepath.Join(repo, "secret.txt"), "safe\nNeed Keyword1 now\nkeyword2 too\nSpee and SeaMy\n")
	mustRun(t, repo, "git", "add", "secret.txt")

	var out bytes.Buffer
	err := runWithOutput([]string{"keyword1", "KEYWORD2", "spee", "seamy"}, &out)
	if !errors.Is(err, errForbiddenWordsFound) {
		t.Fatalf("expected forbidden-word error, got %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "secret.txt:2:Need "+matchColor+"Keyword1"+resetColor+" now") {
		t.Fatalf("missing highlighted keyword1 finding:\n%s", got)
	}
	if !strings.Contains(got, "secret.txt:3:"+matchColor+"keyword2"+resetColor+" too") {
		t.Fatalf("missing highlighted keyword2 finding:\n%s", got)
	}
	if !strings.Contains(got, "secret.txt:4:"+matchColor+"Spee"+resetColor+" and "+matchColor+"SeaMy"+resetColor) {
		t.Fatalf("missing merged highlighted spee/SeaMy finding:\n%s", got)
	}
	if strings.Count(got, "secret.txt:4:") != 1 {
		t.Fatalf("expected one merged finding for two matches on the same line, got:\n%s", got)
	}
}

func TestRunIgnoresDeletedAndUnchangedStagedLines(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	writeFile(t, filepath.Join(repo, "legacy.txt"), "legacy keyword1\nold\nremove keyword2\n")
	mustRun(t, repo, "git", "add", "legacy.txt")
	mustRun(t, repo, "git", "commit", "-m", "init")

	writeFile(t, filepath.Join(repo, "legacy.txt"), "legacy keyword1\nnew safe\n")
	mustRun(t, repo, "git", "add", "legacy.txt")

	var out bytes.Buffer
	if err := runWithOutput([]string{"keyword1", "keyword2"}, &out); err != nil {
		t.Fatalf("expected unchanged/deleted keywords to be ignored, got %v\n%s", err, out.String())
	}
	if out.Len() != 0 {
		t.Fatalf("expected no findings, got:\n%s", out.String())
	}
}

func TestOriginDomainGate(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)
	mustRun(t, repo, "git", "remote", "add", "origin", "git@git.xxx.com:team/repo.git")
	writeFile(t, filepath.Join(repo, "secret.txt"), "keyword1\n")
	mustRun(t, repo, "git", "add", "secret.txt")

	var out bytes.Buffer
	if err := runWithOutput([]string{"--origin-domain", "other.example.com", "keyword1"}, &out); err != nil {
		t.Fatalf("expected mismatched origin domain to skip, got %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("expected no output when origin domain does not match, got:\n%s", out.String())
	}

	err := runWithOutput([]string{"--origin-domain", "git.xxx.com", "keyword1"}, &out)
	if !errors.Is(err, errForbiddenWordsFound) {
		t.Fatalf("expected matching origin domain to scan, got %v", err)
	}
}

func TestExcludeOriginDomainGate(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)
	mustRun(t, repo, "git", "remote", "add", "origin", "git@git.xxx.com:team/repo.git")
	writeFile(t, filepath.Join(repo, "secret.txt"), "keyword1\n")
	mustRun(t, repo, "git", "add", "secret.txt")

	var out bytes.Buffer
	if err := runWithOutput([]string{"--exclude-origin-domain", "git.xxx.com", "keyword1"}, &out); err != nil {
		t.Fatalf("expected matching excluded origin domain to skip, got %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("expected no output when origin domain is excluded, got:\n%s", out.String())
	}

	err := runWithOutput([]string{"--exclude-origin-domain", "other.example.com", "keyword1"}, &out)
	if !errors.Is(err, errForbiddenWordsFound) {
		t.Fatalf("expected non-excluded origin domain to scan, got %v", err)
	}
}

func TestOriginHost(t *testing.T) {
	tests := map[string]string{
		"https://git.xxx.com/team/repo.git":      "git.xxx.com",
		"ssh://git@git.xxx.com:2222/team/repo":   "git.xxx.com",
		"git@git.xxx.com:team/repo.git":          "git.xxx.com",
		"git.xxx.com:team/repo.git":              "git.xxx.com",
		"/Users/me/src/repo":                     "",
		"https://git.xxx.com:8443/team/repo.git": "git.xxx.com",
	}
	for remote, want := range tests {
		if got := githook.OriginHost(remote); got != want {
			t.Fatalf("OriginHost(%q) = %q, want %q", remote, got, want)
		}
	}
}

func initGitRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	mustRun(t, repo, "git", "init")
	mustRun(t, repo, "git", "config", "user.email", "test@example.com")
	mustRun(t, repo, "git", "config", "user.name", "Test User")
	return repo
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func mustRun(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, output)
	}
}
