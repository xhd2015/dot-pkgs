# Scenario

**Feature**: `RunCLI` parses flags, scans repos, formats stdout/stderr

```
# argv -> RunCLI -> Scan -> stdout (lines or JSON) / stderr (errors)
caller Args -> RunCLI -> less-flags parse -> Scan -> format output
```

## Preconditions

- Package `github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo` is importable.
- `RunCLI` is not yet implemented (tests expect RED).
- Fake `.git` fixtures suffice for discovery tests; enrichment leaves need `git` on PATH.

## Context

- `Setup` populates `req.Args` with CLI argv; assertions parse `--root` values via `rootsFromArgs`.
- Paths in assertions use `absPath` for portability.
- Enrichment leaves skip when `git` is unavailable.

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
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

func rootsFromArgs(args []string) []string {
	var roots []string
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--root" && i+1 < len(args):
			roots = append(roots, args[i+1])
			i++
		case strings.HasPrefix(args[i], "--root="):
			roots = append(roots, strings.TrimPrefix(args[i], "--root="))
		}
	}
	return roots
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

func fakeGitRepo(t *testing.T, dir string) {
	t.Helper()
	gitDir := filepath.Join(dir, ".git")
	mkdirAll(t, filepath.Join(gitDir, "objects"))
}

func fakeGitWorktree(t *testing.T, mainDir, wtDir string) {
	t.Helper()
	fakeGitRepo(t, mainDir)
	wtName := filepath.Base(wtDir)
	wtGitDir := filepath.Join(mainDir, ".git", "worktrees", wtName)
	mkdirAll(t, wtGitDir)
	absWtGitDir := absPath(t, wtGitDir)
	writeFile(t, filepath.Join(wtDir, ".git"), "gitdir: "+absWtGitDir+"\n")
}

func cloudStorageProvider(t *testing.T, root, provider string) string {
	t.Helper()
	dir := filepath.Join(root, "Library", "CloudStorage", provider)
	mkdirAll(t, dir)
	return dir
}

func addUnreadableDir(t *testing.T, root, name string) {
	t.Helper()
	secret := filepath.Join(root, name)
	mkdirAll(t, secret)
	if err := os.Chmod(secret, 0000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(secret, 0755)
	})
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Args == nil {
		req.Args = []string{}
	}
	return nil
}
```