## Expected

- The command exits successfully.
- The command reports that `.github/workflows/test.yml` was updated.

## Side Effects

- The stale workflow file is replaced with the generated workflow.

## Exit Code

- Exit code is `0`.

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0\n%s", resp.ExitCode, resp.Output)
	}
	if !strings.Contains(resp.Output, "updated") {
		t.Fatalf("expected updated message, got:\n%s", resp.Output)
	}
	for _, want := range []string{"golang:1.22", "go test -v ./...", "doctest test -v ./..."} {
		if !strings.Contains(resp.WorkflowContent, want) {
			t.Fatalf("expected workflow to contain %q, got:\n%s", want, resp.WorkflowContent)
		}
	}
}
```
