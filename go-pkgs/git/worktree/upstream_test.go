package worktree

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestResolveBranchUpstream_ConfiguredUpstream(t *testing.T) {
	repo := initTempRepo(t, "master")
	runGitT(t, repo, "remote", "add", "origin", "https://example.com/repo.git")
	runGitT(t, repo, "config", "branch.master.remote", "origin")
	runGitT(t, repo, "config", "branch.master.merge", "refs/heads/master")

	remote, remoteBranch, err := ResolveBranchUpstream(repo, "master")
	if err != nil {
		t.Fatal(err)
	}
	if remote != "origin" || remoteBranch != "master" {
		t.Fatalf("got %s %s", remote, remoteBranch)
	}
}

func TestResolveBranchUpstream_FallbackOriginSameName(t *testing.T) {
	repo := initTempRepo(t, "dev-master")
	runGitT(t, repo, "remote", "add", "origin", "https://example.com/repo.git")

	remote, remoteBranch, err := ResolveBranchUpstream(repo, "dev-master")
	if err != nil {
		t.Fatal(err)
	}
	if remote != "origin" || remoteBranch != "dev-master" {
		t.Fatalf("got %s %s", remote, remoteBranch)
	}
}

func TestResolveBranchUpstream_NoRemote(t *testing.T) {
	repo := initTempRepo(t, "master")
	_, _, err := ResolveBranchUpstream(repo, "master")
	if !errors.Is(err, errNoRemoteSync) {
		t.Fatalf("want errNoRemoteSync, got %v", err)
	}
}

func TestResolveBranchUpstream_Detached(t *testing.T) {
	_, _, err := ResolveBranchUpstream("/tmp", "HEAD")
	if err == nil {
		t.Fatal("expected error")
	}
}

func initTempRepo(t *testing.T, branch string) string {
	t.Helper()
	dir := t.TempDir()
	runGitT(t, dir, "init", "-b", branch)
	runGitT(t, dir, "config", "user.email", "test@test.com")
	runGitT(t, dir, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("x\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGitT(t, dir, "add", "README.md")
	runGitT(t, dir, "commit", "-m", "init")
	return dir
}

func runGitT(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
