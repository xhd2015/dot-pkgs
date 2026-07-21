# Scenario

**Feature**: git-hook-no-commit-pattern rejects staged files matching glob patterns

```
# temp git repo with staged files
hook binary <pattern...> -> staged file names -> glob match -> output / exit code
```

## Preconditions

- A temporary git repository exists for each test case.
- The hook command package lives one directory above this doctest tree.
- The `origin` remote is set to `git@github.com:owner/repo.git` by default.

## Steps

1. Initialize a temporary git repository.
2. Configure `origin` remote.
3. Build `git-hook-no-commit-pattern` from its command module into the temporary test directory.
4. Create and stage files as specified by the leaf case.
5. Run the built hook binary with the arguments selected by the leaf case.
6. Capture combined stdout+stderr and the process exit code.

## Context

- The hook checks only staged files (added, copied, modified, renamed, type-changed).
- Deleted staged files are excluded from matching.
- `--origin-domain DOMAIN` limits the hook to repositories whose origin host matches `DOMAIN`.
- `--exclude-origin-domain DOMAIN` skips repositories whose origin host matches `DOMAIN`.
- With no origin filter, the hook evaluates every repository.
- Matching file paths are printed immediately to stdout as they are found.

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CommandDir = filepath.Dir(d.DOCTEST_ROOT)
	req.RepoDir = t.TempDir()
	req.ToolPath = filepath.Join(req.RepoDir, "git-hook-no-commit-pattern")
	build := exec.Command("go", "build", "-o", req.ToolPath, ".")
	build.Dir = req.CommandDir
	if output, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("go build: %w: %s", err, strings.TrimSpace(string(output)))
	}
	if err := runGit(req.RepoDir, "init"); err != nil {
		return err
	}
	if err := runGit(req.RepoDir, "remote", "add", "origin", "git@github.com:owner/repo.git"); err != nil {
		return err
	}
	return nil
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) == 0 {
			return err
		}
		return fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return nil
}

func writeAndStage(root string, name string, content string) error {
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	return runGit(root, "add", name)
}

func deleteAndStage(root string, name string) error {
	path := filepath.Join(root, name)
	if err := os.Remove(path); err != nil {
		return err
	}
	return runGit(root, "add", name)
}
```
