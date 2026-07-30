# Scenario

**Feature**: shared helpers for go-pkgs worktree surface after gitops shim

```
# leaf Setup builds temp main repo (+ optional linked worktrees)
leaf Setup -> git init/commit/worktree add -> req.Dir, paths
# Run calls go-pkgs List / WorktreesOnBranch / IsClean / IsCleanWrk
```

## Preconditions

- `git` available in PATH
- Package `github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree` importable
- Classic TDD: `WorktreesOnBranch` may be missing until P2 re-export (compile RED OK)

## Steps

1. Root verifies `git` is on PATH (skip if missing).
2. Leaves create their own temp repositories via helpers.
3. Helpers provide init, commit, worktree add, path canon, and path membership.

## Context

- Module: `github.com/xhd2015/dot-pkgs/go-pkgs`
- Package under test: `github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree` (not gitops)
- Existing merge-back doctests are separate trees; implementer re-runs them after shim

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = req
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
	}
	return nil
}

func initRepo(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "gopkgs-gitops-shim-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	runGit(t, dir, "init")
	runGit(t, dir, "branch", "-M", "master")
	runGit(t, dir, "config", "user.email", "test@test.com")
	runGit(t, dir, "config", "user.name", "test")
	writeFile(t, dir, "README.md", "init\n")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-m", "init")
	return canonPath(t, dir)
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}

func addLinkedBranch(t *testing.T, mainDir, branch string) string {
	t.Helper()
	linked, err := os.MkdirTemp("", "gopkgs-wt-linked-*")
	if err != nil {
		t.Fatal(err)
	}
	// worktree add requires non-existent path; MkdirTemp created it — remove first.
	if err := os.RemoveAll(linked); err != nil {
		t.Fatal(err)
	}
	runGit(t, mainDir, "worktree", "add", "-b", branch, linked, "HEAD")
	return canonPath(t, linked)
}

func addLinkedExistingBranch(t *testing.T, mainDir, branch string, force bool) string {
	t.Helper()
	linked, err := os.MkdirTemp("", "gopkgs-wt-linked-*")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(linked); err != nil {
		t.Fatal(err)
	}
	args := []string{"worktree", "add"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, linked, branch)
	runGit(t, mainDir, args...)
	return canonPath(t, linked)
}

func canonPath(t *testing.T, path string) string {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return resolved
	}
	dir, base := filepath.Dir(abs), filepath.Base(abs)
	if resolvedDir, err := filepath.EvalSymlinks(dir); err == nil {
		return filepath.Join(resolvedDir, base)
	}
	return filepath.Clean(abs)
}

func samePath(t *testing.T, a, b string) bool {
	t.Helper()
	return canonPath(t, a) == canonPath(t, b)
}

func entryPaths(entries []worktree.Entry) []string {
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.Path
	}
	return out
}

func containsPath(t *testing.T, entries []worktree.Entry, want string) bool {
	t.Helper()
	for _, e := range entries {
		if samePath(t, e.Path, want) {
			return true
		}
	}
	return false
}

func findByPath(t *testing.T, entries []worktree.Entry, want string) (worktree.Entry, bool) {
	t.Helper()
	for _, e := range entries {
		if samePath(t, e.Path, want) {
			return e, true
		}
	}
	return worktree.Entry{}, false
}
```
