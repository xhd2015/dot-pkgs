## Expected

- Exit code 0.
- Linked worktree `Master:` value `identical` is wrapped in green ANSI.
- Stderr is empty.

## Exit Code

- 0

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	if got := statusOutputBlockCount(resp.Stdout); got != 2 {
		t.Fatalf("expected 2 status blocks, got %d:\n%s", got, resp.Stdout)
	}

	root := colorStatusBlockTemplate(t, req.MainRepo, ".", "<ansi-color green>clean</ansi-color>", "")
	assert.Output(t, resp.Stdout, root)

	master := colorStatusMasterFieldColored(t, req.MainRepo, "main", req.WtBranch)
	linked := colorStatusBlockTemplate(t, req.WtDir, "wt-linked", "<ansi-color green>clean</ansi-color>", master)
	assert.Output(t, resp.Stdout, linked)
}
```