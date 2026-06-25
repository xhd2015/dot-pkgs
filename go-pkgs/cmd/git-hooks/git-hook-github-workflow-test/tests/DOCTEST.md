# git-hook-github-workflow-test — Doc-Style Test Tree

## Version
0.0.2

Tests for the GitHub workflow git hook: check mode warns when `.github/workflows/test.yml`
is missing; fix mode creates it for GitHub origins.

# DSN (Domain Specific Notion)

- **Hook binary** — built from the command module, run inside a temp git repository.
- **Repository** — temp git repo with `origin` remote and minimal `go.mod`.
- **Check mode** — warns on missing workflow for `github.com` origins.
- **Fix mode** — creates workflow for GitHub origins; errors on non-GitHub origins.

## How to Run

```sh
doctest test -v ./
```

## Test Tree

- `check/github-missing-workflow`: GitHub origin, no workflow, check mode warns and recommends `--fix`.
- `check/github-existing-workflow`: GitHub origin, existing workflow, check mode exits silently.
- `check/non-github-origin`: Non-GitHub origin, check mode skips silently.
- `check/origin-domain-mismatch`: GitHub origin but `--origin-domain` mismatch, check mode skips silently.
- `fix/github-create-workflow`: GitHub origin, no workflow, `--fix` creates a Go test workflow.
- `fix/github-create-workflow-from-subdir`: GitHub origin, `--fix` run from a subdirectory creates workflow at git toplevel.
- `check/github-existing-workflow-from-subdir`: GitHub origin, workflow at git toplevel, check mode run from subdirectory exits silently.
- `fix/github-existing-workflow`: GitHub origin, existing workflow, `--fix` reports that nothing changed.
- `fix/non-github-origin`: Non-GitHub origin, `--fix` errors and does not create a workflow.
- `fix/origin-domain-mismatch`: GitHub origin but `--origin-domain` mismatch, `--fix` skips silently.
- `args/help`: `--help` prints usage.
- `args/unknown-flag`: unknown flags return an error.

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
	RunDir     string
	ToolPath   string
	RepoDir    string
	Args       []string
	CaseName   string
}

type Response struct {
	Output          string
	ExitCode        int
	WorkflowPath    string
	WorkflowExists  bool
	WorkflowContent string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	cmd := exec.Command(req.ToolPath, req.Args...)
	runDir := req.RepoDir
	if req.RunDir != "" {
		runDir = req.RunDir
	}
	cmd.Dir = runDir
	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, err
		}
	}
	workflowPath := filepath.Join(req.RepoDir, ".github", "workflows", "test.yml")
	content, readErr := os.ReadFile(workflowPath)
	exists := readErr == nil
	if readErr != nil && !os.IsNotExist(readErr) {
		return nil, readErr
	}
	return &Response{
		Output:          string(output),
		ExitCode:        exitCode,
		WorkflowPath:    workflowPath,
		WorkflowExists:  exists,
		WorkflowContent: string(content),
	}, nil
}
```