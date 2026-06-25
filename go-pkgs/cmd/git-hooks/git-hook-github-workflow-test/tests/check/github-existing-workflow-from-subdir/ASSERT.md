## Expected

- The command exits successfully.
- The command produces no output because the workflow already exists at the git repository root.

## Side Effects

- The workflow file at the git repository root remains unchanged.

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
	if strings.TrimSpace(resp.Output) != "" {
		t.Fatalf("expected silent success when workflow exists at git toplevel, got:\n%s", resp.Output)
	}
	if !resp.WorkflowExists {
		t.Fatalf("expected workflow to remain at git toplevel %s", resp.WorkflowPath)
	}
}
```