## Expected

- Non-zero exit code.
- Empty stdout preferred.
- Stderr mentions unknown var (`no_such`) or template/expand error.

## Errors

- Static check or dry expand rejects unknown `${name}`.

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
		t.Fatalf("expected non-zero for unknown var, stdout=%q", resp.Stdout)
	}
	assertEmptyStdout(t, resp.Stdout)
	assert.Output(t, resp.Stderr, `<contains>
no_such
</contains>`)
}
```
