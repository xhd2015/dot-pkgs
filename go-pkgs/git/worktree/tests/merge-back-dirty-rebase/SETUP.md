# Scenario

**Feature**: merge-back with dirty diverged worktrees uses tmp worktree for rebase

```
# Setup chain: init git fixture → build branch topology → run MergeBack → assert
Setup -> git repo + linked worktree -> MergeBack(opts) -> result + checkout state
```

## Preconditions

- `git` available on PATH.
- Package `github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree` importable.

## Context

- Each test uses `t.TempDir()` as `WorkRoot`.
- `req.MakeDirty` controls whether the source worktree has uncommitted changes.
- Branch topology is built by descendant SETUPs.
- Tests assert both the `MergeBackResult` (Action, Relation) and filesystem side-effects (tmp worktree cleanup, branch state).

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func skipIfNoGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
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

func runGitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if exitErr != nil {
			t.Fatalf("git %v in %s: %v\n%s", args, dir, err, exitErr.Stderr)
		}
		t.Fatalf("git %v in %s: %v", args, dir, err)
	}
	return strings.TrimSpace(string(out))
}

func initRepo(t *testing.T, path, defaultBranch string) {
	t.Helper()
	if defaultBranch != "" {
		runGit(t, path, "init", "-b", defaultBranch)
	} else {
		runGit(t, path, "init")
	}
	runGit(t, path, "config", "user.email", "test@test.com")
	runGit(t, path, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(path, "README.md"), []byte("# test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, path, "add", "README.md")
	runGit(t, path, "commit", "-m", "init")
}

func addWorktree(t *testing.T, mainRepo, wtDir, branch string) {
	t.Helper()
	runGit(t, mainRepo, "worktree", "add", "-b", branch, wtDir)
}

func makeDirty(t *testing.T, dir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "dirty.txt"), []byte("uncommitted\n"), 0644); err != nil {
		t.Fatal(err)
	}
}

func revParseHEAD(t *testing.T, dir string) string {
	t.Helper()
	return runGitOutput(t, dir, "rev-parse", "HEAD")
}

func isClean(t *testing.T, dir string) bool {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git status in %s: %v", dir, err)
	}
	return len(strings.TrimSpace(string(out))) == 0
}

func hasBranch(t *testing.T, dir string, branch string) bool {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--verify", "refs/heads/"+branch)
	err := cmd.Run()
	return err == nil
}

func hasDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func branchCommit(t *testing.T, dir string, branch string) string {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--verify", "refs/heads/"+branch)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("rev-parse %s in %s: %v", branch, dir, err)
	}
	return strings.TrimSpace(string(out))
}

func readDefaultBranch(t *testing.T, repo string) string {
	t.Helper()
	cmd := exec.Command("git", "-C", repo, "branch", "--show-current")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("read default branch: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func isAncestor(t *testing.T, repo, ancestor, descendant string) bool {
	t.Helper()
	cmd := exec.Command("git", "-C", repo, "merge-base", "--is-ancestor", ancestor, descendant)
	err := cmd.Run()
	return err == nil
}

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	req.WorkRoot = t.TempDir()
	return nil
}
```
