# Scenario

**Feature**: shared helpers for `worktree.ReadBranch`

```
leaf Setup -> git init [-b main] [commit] [detach] -> req.Dir
Run -> worktree.ReadBranch(req.Dir) -> Response.Branch
```

## Preconditions

- `git` available in PATH (skip otherwise).
- Package `github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree` importable.

## Steps

1. Root verifies `git` is on PATH.
2. Leaves create isolated temp repos via helpers (`-b main`, hooks disabled).

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	_ = req
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
	}
	return nil
}

func initUnbornRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "--template=", "-b", "main")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")
	runGit(t, dir, "config", "core.hooksPath", "/dev/null")
	return dir
}

func firstCommit(t *testing.T, dir string) {
	t.Helper()
	path := filepath.Join(dir, "README.md")
	if err := os.WriteFile(path, []byte("init\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "README.md")
	runGit(t, dir, "commit", "-m", "init")
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL=/dev/null",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}
```
