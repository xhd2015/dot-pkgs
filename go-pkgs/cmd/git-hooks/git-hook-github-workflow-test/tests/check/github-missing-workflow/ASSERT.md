## Expected

- The command exits successfully.
- The output warns that `.github/workflows/test.yml` is missing.
- The output recommends `git-hook-github-workflow-test --fix`.

## Side Effects

- No workflow file is created in check mode.

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
		t.Fatalf("expected missing workflow warning, got:\n%s", resp.Output)
	}
	if !strings.Contains(resp.Output, "git-hook-github-workflow-test --fix") {
		t.Fatalf("expected fix recommendation, got:\n%s", resp.Output)
	}
	if resp.WorkflowExists {
		t.Fatalf("check mode must not create workflow at %s", resp.WorkflowPath)
	}
}
```
