# Scenario

**Feature**: `--auto-unstage` flag unstages matched files instead of failing

```
# hook with --auto-unstage and patterns -> match -> print + unstage -> exit 0
hook binary --auto-unstage <patterns> -> matches found -> print paths -> git restore --staged -> exit 0

# non-matched staged files remain untouched
git restore --staged -- <matched files> -> matched files removed from index, working copy preserved
```

## Preconditions

- A git repository with an initial commit is required (so `git restore --staged` can distinguish staged from committed).
- The hook binary is built by the root SETUP.

## Steps

1. The root SETUP builds the binary and initializes the repo.
2. Each leaf creates an initial commit, then stages files and sets `req.Args`.
3. After the hook runs, verify staged state via `git diff --cached`.

## Context

- `--auto-unstage` calls `git restore --staged -- <files>` on matched paths.
- The hook returns exit code 0 (not errPatternsMatched) when `--auto-unstage` is active.
- Working copy files are not removed — only the index is modified.

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = nil // leaf cases set args
	return nil
}

func initGitRepoWithCommit(repoDir string) error {
	if err := writeAndStage(repoDir, ".gitkeep", ""); err != nil {
		return err
	}
	return runGit(repoDir, "commit", "--no-verify", "-m", "initial")
}

func getStagedFileNames(repoDir string) ([]string, error) {
	cmd := exec.Command("git", "-C", repoDir, "diff", "--cached", "--name-only", "--diff-filter=ACMRT")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func containsString(items []string, s string) bool {
	for _, item := range items {
		if item == s {
			return true
		}
	}
	return false
}
```
