## Expected

- Non-zero exit code.
- Empty stdout preferred.
- Stderr mentions missing/absent interceptor or not configured.

## Errors

- Dry-run requires a present enabled interceptor.

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
		t.Fatalf("expected non-zero when interceptor absent")
	}
	assertEmptyStdout(t, resp.Stdout)
	assert.Output(t, resp.Stderr, `<contains>
interceptor
</contains>`)
}
```
