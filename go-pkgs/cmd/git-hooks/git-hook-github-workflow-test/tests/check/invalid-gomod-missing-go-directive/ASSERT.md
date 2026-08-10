## Expected

- The command exits successfully.
- The output warns that `go.mod` is missing a `go` directive.
- The output still warns that `.github/workflows/test.yml` is missing.

## Side Effects

- No workflow file is created in check mode.

## Exit Code

- Exit code is `0`.

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0\n%s", resp.ExitCode, resp.Output)
	}
	if !strings.Contains(resp.Output, "warning:") || !strings.Contains(resp.Output, "missing go directive") {
		t.Fatalf("expected missing go directive warning, got:\n%s", resp.Output)
	}
	if !strings.Contains(resp.Output, ".github/workflows/test.yml") {
		t.Fatalf("expected missing workflow warning, got:\n%s", resp.Output)
	}
	if resp.WorkflowExists {
		t.Fatalf("check mode must not create workflow at %s", resp.WorkflowPath)
	}
}
```
