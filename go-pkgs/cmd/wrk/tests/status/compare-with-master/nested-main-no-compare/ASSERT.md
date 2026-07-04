## Expected

- Exit code 0.
- Two status blocks for `.` and `tools/child`.
- Neither block contains `Compare with Master`.
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
	assert.Output(t, resp.Stdout, statusBlockTemplate(t, req.DepPath, "tools/child", "clean"))
	assertNoCompareWithMaster(t, resp.Stdout)
}
```