# Scenario

**Feature**: merge-back plan display uses comments, Short paths, and dynamic target labels

```
# library doctest harness builds git fixture then MergeBack
Setup -> git repo + linked worktree -> Run(MergeBackOptions) -> captured prompt or dry-run stdout
```

## Preconditions

- `git` available on PATH.
- Package `github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree` importable.

## Context

- Each test uses `t.TempDir()` as `WorkRoot`.
- `DefaultBranch` empty → plain `git init`; non-empty → `git init -b <name>`.
- Display assertions use `shortPath` (mirrors merge-back display: `/private/var`
  normalization + `pathfmt.Short` for home-relative paths).

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/pathfmt"
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

func readDefaultBranch(t *testing.T, repo string) string {
	t.Helper()
	cmd := exec.Command("git", "-C", repo, "branch", "--show-current")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("read default branch: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func addWorktree(t *testing.T, mainRepo, wtDir, branch string) {
	t.Helper()
	runGit(t, mainRepo, "worktree", "add", "-b", branch, wtDir)
}

func revParseHEAD(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git rev-parse HEAD in %s: %v", dir, err)
	}
	return strings.TrimSpace(string(out))
}

func shortPath(t *testing.T, abs string) string {
	t.Helper()
	p := filepath.Clean(abs)
	if strings.HasPrefix(p, "/private/var/") {
		p = "/var/" + strings.TrimPrefix(p, "/private/var/")
	}
	return pathfmt.Short(p)
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	skipIfNoGit(t)
	req.WorkRoot = t.TempDir()
	return nil
}
```