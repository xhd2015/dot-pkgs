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

func TestNonWipCommitPasses(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	doCommit(t, repo, "feat: add feature")
	writeFile(t, filepath.Join(repo, "x.txt"), "safe\n")
	mustRun(t, repo, "git", "add", "x.txt")

	var out bytes.Buffer
	err := runWithOutput([]string{}, &out)
	if err != nil {
		t.Fatalf("expected non-WIP commit to pass, got %v\n%s", err, out.String())
	}
}

func TestNonWipCommitPassesPush(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	doCommit(t, repo, "feat: add feature")

	var out bytes.Buffer
	err := runWithOutput([]string{"--phase", "push"}, &out)
	if err != nil {
		t.Fatalf("expected non-WIP commit to pass push, got %v\n%s", err, out.String())
	}
}

func TestWipColonPrefixRejected(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	doCommit(t, repo, "WIP: refactor module")
	writeFile(t, filepath.Join(repo, "a.txt"), "change\n")
	mustRun(t, repo, "git", "add", "a.txt")

	var out bytes.Buffer
	err := runWithOutput([]string{}, &out)
	if !errors.Is(err, errWipProtected) {
		t.Fatalf("expected WIP rejected, got %v\n%s", err, out.String())
	}
}

func TestWipLowercaseColonRejected(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	doCommit(t, repo, "wip: draft work")
	writeFile(t, filepath.Join(repo, "b.txt"), "change\n")
	mustRun(t, repo, "git", "add", "b.txt")

	var out bytes.Buffer
	err := runWithOutput([]string{}, &out)
	if !errors.Is(err, errWipProtected) {
		t.Fatalf("expected lowercase wip: rejected, got %v\n%s", err, out.String())
	}
}

func TestWipExactRejected(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	doCommit(t, repo, "WIP")
	writeFile(t, filepath.Join(repo, "c.txt"), "change\n")
	mustRun(t, repo, "git", "add", "c.txt")

	var out bytes.Buffer
	err := runWithOutput([]string{}, &out)
	if !errors.Is(err, errWipProtected) {
		t.Fatalf("expected exact WIP rejected, got %v\n%s", err, out.String())
	}
}

func TestWipLowercaseExactRejected(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	doCommit(t, repo, "wip")
	writeFile(t, filepath.Join(repo, "d.txt"), "change\n")
	mustRun(t, repo, "git", "add", "d.txt")

	var out bytes.Buffer
	err := runWithOutput([]string{}, &out)
	if !errors.Is(err, errWipProtected) {
		t.Fatalf("expected lowercase wip rejected, got %v\n%s", err, out.String())
	}
}

func TestPreCommitAmendAllowsWip(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	doCommit(t, repo, "WIP: in progress")
	writeFile(t, filepath.Join(repo, "e.txt"), "change\n")
	mustRun(t, repo, "git", "add", "e.txt")

	var out bytes.Buffer
		err := runWithOutput([]string{"--is-amend"}, &out)
	if err != nil {
		t.Fatalf("expected --is-amend to bypass WIP check, got %v\n%s", err, out.String())
	}
}

func TestPushRejectsWipEvenWithAmend(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	doCommit(t, repo, "WIP: in progress")

	var out bytes.Buffer
	err := runWithOutput([]string{"--phase", "push", "--is-amend"}, &out)
	if !errors.Is(err, errWipProtected) {
		t.Fatalf("expected push to reject WIP even with --is-amend, got %v\n%s", err, out.String())
	}
}

func TestPushRejectsWipExact(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	doCommit(t, repo, "wip")

	var out bytes.Buffer
	err := runWithOutput([]string{"--phase", "push"}, &out)
	if !errors.Is(err, errWipProtected) {
		t.Fatalf("expected push to reject WIP, got %v\n%s", err, out.String())
	}
}

func TestWipWithinWordNotRejected(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	doCommit(t, repo, "fix: equipper alignment")
	writeFile(t, filepath.Join(repo, "f.txt"), "change\n")
	mustRun(t, repo, "git", "add", "f.txt")

	var out bytes.Buffer
	err := runWithOutput([]string{}, &out)
	if err != nil {
		t.Fatalf("expected non-WIP commit with 'wip' inside word to pass, got %v\n%s", err, out.String())
	}
}

func TestOriginDomainGate(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)
	mustRun(t, repo, "git", "remote", "add", "origin", "git@git.xxx.com:team/repo.git")
	doCommit(t, repo, "WIP: draft")

	writeFile(t, filepath.Join(repo, "g.txt"), "change\n")
	mustRun(t, repo, "git", "add", "g.txt")

	var out bytes.Buffer
	if err := runWithOutput([]string{"--origin-domain", "other.example.com"}, &out); err != nil {
		t.Fatalf("expected mismatched origin domain to skip, got %v\n%s", err, out.String())
	}
	if out.Len() != 0 {
		t.Fatalf("expected no output when origin domain does not match, got:\n%s", out.String())
	}

	err := runWithOutput([]string{"--origin-domain", "git.xxx.com"}, &out)
	if !errors.Is(err, errWipProtected) {
		t.Fatalf("expected matching origin domain to detect WIP, got %v\n%s", err, out.String())
	}
}

func TestExcludeOriginDomainGate(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)
	mustRun(t, repo, "git", "remote", "add", "origin", "git@git.xxx.com:team/repo.git")
	doCommit(t, repo, "WIP: draft")

	writeFile(t, filepath.Join(repo, "h.txt"), "change\n")
	mustRun(t, repo, "git", "add", "h.txt")

	var out bytes.Buffer
	if err := runWithOutput([]string{"--exclude-origin-domain", "git.xxx.com"}, &out); err != nil {
		t.Fatalf("expected excluded origin domain to skip, got %v\n%s", err, out.String())
	}
	if out.Len() != 0 {
		t.Fatalf("expected no output when origin domain is excluded, got:\n%s", out.String())
	}

	err := runWithOutput([]string{"--exclude-origin-domain", "other.example.com"}, &out)
	if !errors.Is(err, errWipProtected) {
		t.Fatalf("expected non-excluded origin domain to detect WIP, got %v\n%s", err, out.String())
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

func TestNoCommitsYetPasses(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	var out bytes.Buffer
	err := runWithOutput([]string{}, &out)
	if err != nil {
		t.Fatalf("expected no-commits repo to pass, got %v\n%s", err, out.String())
	}
}

func TestPhaseFromEnv(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	doCommit(t, repo, "WIP: draft")

	t.Setenv("GIT_HOOK_PHASE", "push")
	var out bytes.Buffer
	err := runWithOutput([]string{}, &out)
	if !errors.Is(err, errWipProtected) {
		t.Fatalf("expected push phase from GIT_HOOK_PHASE env to reject WIP, got %v\n%s", err, out.String())
	}
}

func doCommit(t *testing.T, repo string, msg string) {
	t.Helper()
	writeFile(t, filepath.Join(repo, ".keep"), "")
	mustRun(t, repo, "git", "add", ".keep")
	mustRun(t, repo, "git", "commit", "-m", msg)
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
