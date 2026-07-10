## Expected

- Non-zero exit code (prefer 1).
- Stderr contains a useful message about missing interceptor / config.
- Stdout empty (or not a successful JSON dump).

## Errors

- No interceptor section to show.

## Exit Code

- non-zero

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q", resp.Stdout)
	}
	assertEmptyStdout(t, resp.Stdout)
	assert.Output(t, resp.Stderr, `<contains>
interceptor
</contains>`)
}
```
