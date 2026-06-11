## Expected

- The command exits successfully.
- The command warns that `.github/workflows/test.yml` differs from the recommended workflow.
- The output recommends `git-hook-github-workflow-test --fix`.

## Side Effects

- The existing workflow file remains unchanged.

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
	if !strings.Contains(resp.Output, ".github/workflows/test.yml") {
		t.Fatalf("expected workflow warning, got:\n%s", resp.Output)
	}
	if !strings.Contains(resp.Output, "differs") {
		t.Fatalf("expected differs warning, got:\n%s", resp.Output)
	}
	if !strings.Contains(resp.Output, "git-hook-github-workflow-test --fix") {
		t.Fatalf("expected fix recommendation, got:\n%s", resp.Output)
	}
	if resp.WorkflowContent != "name: existing\n" {
		t.Fatalf("workflow changed unexpectedly:\n%s", resp.WorkflowContent)
	}
}
```
