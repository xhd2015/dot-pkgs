## Expected

- Non-zero exit code.
- Empty stdout preferred.
- Stderr mentions disabled / not enabled.

## Errors

- Dry-run requires enabled interceptor (no dry-run-when-disabled flag in v1).

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
		t.Fatalf("expected non-zero when interceptor disabled")
	}
	assertEmptyStdout(t, resp.Stdout)
	assert.Output(t, resp.Stderr, `<contains>
enabl
</contains>`)
}
```
