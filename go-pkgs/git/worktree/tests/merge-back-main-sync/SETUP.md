# Scenario

**Feature**: MergeBack syncs main onto upstream before land

```
Setup -> WorkRoot temp dir + git helpers
```

## Preconditions

- `git` on PATH

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
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
		t.Fatalf("git %v in %s: %v", args, dir, err)
	}
	return strings.TrimSpace(string(out))
}

func initRepo(t *testing.T, path, branch string) {
	t.Helper()
	runGit(t, path, "init", "-b", branch)
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

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	skipIfNoGit(t)
	workRoot, err := os.MkdirTemp("", "gopkgs-merge-back-main-sync-*")
	if err != nil {
		return err
	}
	t.Cleanup(func() { _ = os.RemoveAll(workRoot) })
	req.WorkRoot = workRoot
	return nil
}
```
