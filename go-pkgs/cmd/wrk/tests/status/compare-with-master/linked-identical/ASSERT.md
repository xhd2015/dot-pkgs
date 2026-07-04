## Expected

- Exit code 0.
- Linked worktree block includes `Compare with Master: main and wt-side are identical`.
- Root block has no `Compare with Master` line.
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

	assert.Output(t, resp.Stdout, statusBlockTemplate(t, req.MainRepo, ".", "clean"))

	compare := compareWithMasterField(t, req.MainRepo, "main", req.WtBranch)
	linkedBlock := statusBlockWithCompare(t, req.WtDir, "wt-linked", "clean", compare)
	assert.Output(t, resp.Stdout, linkedBlock)
}
```