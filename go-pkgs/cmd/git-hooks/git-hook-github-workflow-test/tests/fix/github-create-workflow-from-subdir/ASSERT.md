## Expected

- The command exits successfully.
- `.github/workflows/test.yml` is created at the git repository root, not in the subdirectory.
- The workflow uses container image `golang:1.22`.
- The workflow runs `go test -v ./...`.
- The workflow runs `doctest test -v ./...`.

## Side Effects

- No workflow file is created under the subdirectory working directory.

## Exit Code

- Exit code is `0`.

```go
import (
	"os"
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0\n%s", resp.ExitCode, resp.Output)
	}
	if !resp.WorkflowExists {
		t.Fatalf("expected workflow to be created at git toplevel %s", resp.WorkflowPath)
	}
	subdirWorkflow := filepath.Join(req.RunDir, ".github", "workflows", "test.yml")
	if _, statErr := os.Stat(subdirWorkflow); statErr == nil {
		t.Fatalf("workflow must not be created in subdirectory %s", subdirWorkflow)
	} else if !os.IsNotExist(statErr) {
		t.Fatalf("stat subdirectory workflow: %v", statErr)
	}
	content := resp.WorkflowContent
	for _, want := range []string{"container:", "golang:1.22", "go test -v ./...", "doctest test -v ./..."} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected workflow to contain %q, got:\n%s", want, content)
		}
	}
}
```