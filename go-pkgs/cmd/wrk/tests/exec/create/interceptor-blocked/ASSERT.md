## Expected

- Non-zero exit.
- Stderr mentions interceptor and/or `--no-interceptor` (user must disable intercept to use `--exec` with create).
- Stdout empty (or no successful path+exec sequence).

## Errors

- Create interceptor cannot be combined with `--exec` unless escape hatch.

## Exit Code

- Non-zero

```go
import (
	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	// Soft check: error must guide the user toward disabling the interceptor.
	assert.Output(t, resp.Stderr, `<contains>
interceptor
</contains>`)
	assert.Output(t, resp.Stderr, `<contains>
--no-interceptor
</contains>`)
}
```
