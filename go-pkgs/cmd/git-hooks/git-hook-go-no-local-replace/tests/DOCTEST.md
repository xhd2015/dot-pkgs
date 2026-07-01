# git-hook-go-no-local-replace — Doc-Style Test Tree

## Version
0.0.2

Tests for the git hook that rejects local `replace` directives in go.mod files.

# DSN (Domain Specific Notion)

- **Hook binary** — built from the command module, run inside a temp git repository.
- **Repository** — temp git repo with optional `origin` remote and one or more go.mod files.
- **Scan** — the hook uses `gotool/mod/scan` to walk all go.mod files under the repo root.
- **Replace detection** — for each module, the hook inspects `replace` directives. A replace is "local" when `NewVersion` is empty (Go modfile convention for path-based replaces such as `./xxx`, `../xxx`, or `/abs/path`). Version-only replaces (`old v1.0.0 => v2.0.0`) are not local.
- **Streaming output** — each local replace is printed to stdout as soon as it is found.
- **Exit code** — 0 when no local replaces are found; 1 when any local replace is found.

## How to Run

```sh
doctest test -v ./
```

## Test Tree

- `args/help`: `--help` prints usage.
- `args/unknown-flag`: unknown flags return an error.
- `domain-gate/origin-domain-mismatch`: `--origin-domain` mismatch skips silently.
- `domain-gate/exclude-origin-domain-match`: `--exclude-origin-domain` match skips silently.
- `scan/no-go-mod`: no go.mod files in repo, exit 0.
- `scan/no-replaces`: go.mod with no replace directives, exit 0.
- `scan/version-replace-only`: go.mod with only version-based replaces, exit 0.
- `scan/local-replace/dot-slash`: go.mod with `./xxx` local replace, prints path, exit 1.
- `scan/local-replace/dot-dot-slash`: go.mod with `../xxx` local replace, prints path, exit 1.
- `scan/local-replace/abs-path`: go.mod with absolute-path local replace, prints path, exit 1.
- `scan/local-replace/multi-module`: multiple go.mod files, some with local replaces, streams each, exit 1.

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
