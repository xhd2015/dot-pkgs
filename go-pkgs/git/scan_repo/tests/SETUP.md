# Scenario

**Feature**: `scan_repo.Scan` discovers git repos; `ParseRemoteOwnerRepo` parses remote URLs

```
# discovery pipeline
caller roots + options -> Scan -> Walk -> repo rows (sorted by Path)

# enrichment (optional)
Scan -> git subprocesses -> Remotes[] / Worktrees[] on main rows

# URL parser (standalone)
ParseRemoteOwnerRepo(url) -> owner, repo, ok
```

## Preconditions

- Package `github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo` is importable.
- `Scan` uses filesystem walk only when enrichment flags are false.
- Enrichment branches require `git` on PATH.

## Context

- Paths in assertions use `filepath.Abs` and `filepath.Clean` for portability.
- `fakeGitRepo` / `fakeGitWorktree` avoid git for pure discovery tests.
- Real git helpers skip the test when `git` is unavailable.

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func absPath(t *testing.T, p string) string {
	t.Helper()
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(abs)
}

func mkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func gitAvailable(t *testing.T) bool {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
		return false
	}
	return true
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}
```