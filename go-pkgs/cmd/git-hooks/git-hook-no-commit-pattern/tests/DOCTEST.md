# git-hook-no-commit-pattern — Doc-Style Test Tree

## Version
0.0.2

Tests for the git hook that rejects staged files whose names match user-supplied glob patterns.

# DSN (Domain Specific Notion)

- **Hook binary** — built from the command module, run inside a temp git repository.
- **Repository** — temp git repo with optional `origin` remote and staged files.
- **Patterns** — glob patterns provided as positional arguments (e.g. `REQUIREMENT-*`, `*.exe`).
- **Staged files** — files added to the git index (`git diff --cached --name-only --diff-filter=ACMRT --`). Deleted files (`D` status) are excluded.
- **Match detection** — each staged file name is checked against all provided patterns. On first match per file, the file path is printed immediately.
- **Exit code** — 0 when no staged files match any pattern; 1 when any match is found.

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

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
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

func Run(t *testing.T, req *Request) (*Response, error) {
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
