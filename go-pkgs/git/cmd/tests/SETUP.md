# Scenario

**Feature**: `git/cmd.Run` wraps git subprocess with normalized errors

```
caller dir + args -> Run -> git subprocess -> trimmed stdout or one-line error
```

## Preconditions

- Package `github.com/xhd2015/dot-pkgs/go-pkgs/git/cmd` is importable.
- Real `git` on PATH for leaves that exercise subprocesses.

## Context

- Tests skip when `git` is unavailable.
- Helpers create minimal repos under `t.TempDir()`.

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	return nil
}

func gitAvailable(t *testing.T) bool {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
		return false
	}
	return true
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

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}

func gitInitRepo(t *testing.T, dir string) {
	t.Helper()
	mkdirAll(t, dir)
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")
}

func gitInitialCommit(t *testing.T, dir string) {
	t.Helper()
	writeFile(t, filepath.Join(dir, "README"), "init\n")
	runGit(t, dir, "add", "README")
	runGit(t, dir, "commit", "-m", "init")
}
```
