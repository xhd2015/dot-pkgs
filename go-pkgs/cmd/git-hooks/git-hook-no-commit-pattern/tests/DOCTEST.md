# git-hook-no-commit-pattern — Doc-Style Test Tree

## Version
0.0.3

Tests for the git hook that rejects staged files whose names match user-supplied glob patterns.

# DSN (Domain Specific Notion)

- **Hook binary** — built from the command module, run inside a temp git repository.
- **Repository** — temp git repo with optional `origin` remote and staged files.
- **Patterns** — glob patterns provided as positional arguments (e.g. `REQUIREMENT-*`, `*.exe`).
- **Staged files** — files added to the git index (`git diff --cached --name-only --diff-filter=ACMRT --`). Deleted files (`D` status) are excluded.
- **Match detection** — each staged path is checked against all provided patterns. Patterns without `/` are matched against every path segment; patterns with `/` are matched against the full staged path. On first match per file, the file path is printed immediately.
- **Exit code** — 0 when no staged files match any pattern; 1 when any match is found.
- **`--auto-unstage`** — when set, matched files are automatically unstaged (via `git restore --staged`) instead of failing the commit. Exit 0, matched paths still printed. Non-matched staged files remain staged.

## How to Run

```sh
doctest test -v ./
```

## Test Tree

- `args/help`: `--help` prints usage.
- `args/no-pattern`: no patterns provided, returns an error.
- `domain-gate/origin-domain-mismatch`: `--origin-domain` mismatch skips silently.
- `domain-gate/exclude-origin-domain-match`: `--exclude-origin-domain` match skips silently.
- `match/no-match`: staged files exist but none match the pattern, exit 0.
- `match/single-match`: one staged file matches the pattern, prints path, exit 1.
- `match/multi-match`: multiple staged files match patterns, prints each, exit 1.
- `match/deleted-excluded`: staged deleted file matches pattern but is excluded, exit 0.
- `match/multi-pattern`: multiple patterns provided, staged file matches one, exit 1.
- `match/requirement-glob-subdir`: `REQUIREMENT-*.md` matches staged requirement docs in subdirectories, exit 1.
- `match/segment-dot-vscode`: slashless `.vscode` matches a middle path segment, exit 1.
- `match/slash-pattern-no-match`: patterns with `/` do not match when only a segment matches, exit 0.
- `auto-unstage/single-match-unstage`: `--auto-unstage` + one match → print, unstage, exit 0.
- `auto-unstage/multi-match-unstage`: `--auto-unstage` + multiple matches → print all, unstage all, exit 0.
- `auto-unstage/partial-unstage`: `--auto-unstage` + mixed staged (match + no-match) → only matched unstaged.
- `auto-unstage/no-match`: `--auto-unstage` + no matches → exit 0, no output, nothing untaged.
- `auto-unstage/domain-gate-skip`: `--auto-unstage` + domain filter skip → no effect.
- `auto-unstage/requirement-glob-subdir`: `--auto-unstage` + `REQUIREMENT-*.md` matches and unstages requirement docs in subdirectories.

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

type Request struct {
	CommandDir string
	ToolPath   string
	RepoDir    string
	Args       []string
}

type Response struct {
	Output   string
	ExitCode int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	cmd := exec.Command(req.ToolPath, req.Args...)
	cmd.Dir = req.RepoDir
	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, err
		}
	}
	return &Response{
		Output:   string(output),
		ExitCode: exitCode,
	}, nil
}
```
