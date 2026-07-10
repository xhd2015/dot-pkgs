## Expected

- Non-zero exit.
- Stderr is exactly `wrk: --cd requires a path argument` (or contains that phrase).
- Stdout empty.

## Errors

- Missing required path for `--cd`.

## Exit Code

- Non-zero

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assertEmptyStdout(t, resp.Stdout)
	assert.Output(t, resp.Stderr, `wrk: --cd requires a path argument`)
}
```
