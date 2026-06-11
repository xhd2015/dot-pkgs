## Expected

- The command exits with an error.
- The output states that `--fix` requires a `github.com` origin.

## Side Effects

- No workflow file is created.

## Exit Code

- Exit code is non-zero.

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit code for non-GitHub fix\n%s", resp.Output)
	}
	if !strings.Contains(resp.Output, "github.com") {
		t.Fatalf("expected github.com error, got:\n%s", resp.Output)
	}
	if resp.WorkflowExists {
		t.Fatalf("non-GitHub fix must not create workflow at %s", resp.WorkflowPath)
	}
}
```
