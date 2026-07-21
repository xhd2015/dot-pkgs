# Scenario

**Feature**: git-hook-go-no-local-replace rejects local replace directives in go.mod

```
# temp git repo with optional go.mod files
hook binary -> scan go.mod -> replace inspection -> output / exit code
```

## Preconditions

- A temporary git repository exists for each test case.
- The hook command package lives one directory above this doctest tree.
- The `origin` remote is set to `git@github.com:owner/repo.git` by default.

## Steps

1. Initialize a temporary git repository.
2. Configure `origin` remote.
3. Build `git-hook-go-no-local-replace` from its command module into the temporary test directory.
4. Write go.mod files and other fixtures as specified by the leaf case.
5. Run the built hook binary with the arguments selected by the leaf case.
6. Capture combined stdout+stderr and the process exit code.

## Context

- The hook scans all go.mod files under the repository root.
- `--origin-domain DOMAIN` limits the hook to repositories whose origin host matches `DOMAIN`.
- `--exclude-origin-domain DOMAIN` skips repositories whose origin host matches `DOMAIN`.
- With no origin filter, the hook evaluates every repository.
- Replace directives with `NewVersion == ""` are local-path replaces.
- Version-only replaces set `NewVersion` to the version string and are not local.
- Local replaces are printed immediately as they are found.

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
	req.ToolPath = filepath.Join(req.RepoDir, "git-hook-go-no-local-replace")
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

func writeGoMod(root string, rel string, content string) error {
	dir := filepath.Dir(filepath.Join(root, rel))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, rel), []byte(content), 0o644)
}

```
