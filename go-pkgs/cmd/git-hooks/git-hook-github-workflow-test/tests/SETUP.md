# Scenario

**Feature**: git-hook-github-workflow-test checks and fixes GitHub CI workflow files

```
# temp git repo with origin + go.mod
hook binary -> check/fix mode -> workflow file state
```

## Preconditions

- A temporary git repository exists for each test case.
- The repository has a `remote.origin.url`.
- The repository contains a `go.mod` file with Go version `1.22`.
- The hook command package lives one directory above this doctest tree.

## Steps

1. Initialize a temporary git repository.
2. Configure `origin` to point at a GitHub repository by default.
3. Write a minimal `go.mod`.
4. Build `git-hook-github-workflow-test` from its command module into the temporary test directory.
5. Run the built hook binary with the arguments selected by the leaf case.
6. Capture stdout and stderr together, the process exit code, and the final workflow file state.

## Context

- The hook checks the current git repository.
- `--origin-domain DOMAIN` limits the hook to repositories whose origin host matches `DOMAIN`.
- `--exclude-origin-domain DOMAIN` skips repositories whose origin host matches `DOMAIN`.
- With no origin filter, the hook evaluates every repository, but it only warns or fixes when the origin host is exactly `github.com`.
- In check mode, a missing `.github/workflows/test.yml` on GitHub prints a warning recommending `git-hook-github-workflow-test --fix`.
- In fix mode, non-GitHub origins are errors. Existing workflow files are not overwritten. Missing workflow files are created.
- Created workflow files use a `golang:<go-version>` container image derived from `go.mod` and run `go test -v ./...` followed by `doctest test -v --label-all ./...`.

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
	req.CommandDir = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, ".."))
	req.RepoDir = t.TempDir()
	req.ToolPath = filepath.Join(req.RepoDir, "git-hook-github-workflow-test")
	req.CaseName = "default"
	build := exec.Command("go", "build", "-o", req.ToolPath, ".")
	build.Dir = req.CommandDir
	if output, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("go build git-hook-github-workflow-test: %w: %s", err, strings.TrimSpace(string(output)))
	}
	if err := runGit(req.RepoDir, "init"); err != nil {
		return err
	}
	if err := runGit(req.RepoDir, "remote", "add", "origin", "git@github.com:owner/repo.git"); err != nil {
		return err
	}
	if err := writeFile(req.RepoDir, "go.mod", "module example.com/repo\n\ngo 1.22\n"); err != nil {
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

func writeFile(root string, rel string, content string) error {
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func setOrigin(req *Request, remote string) error {
	return runGit(req.RepoDir, "remote", "set-url", "origin", remote)
}
```