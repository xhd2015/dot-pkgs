## Expected

- Non-zero exit code.
- Empty stdout preferred.
- Stderr indicates a required action or usage for interceptor management.

## Errors

- Exactly one of `--status|--show|--path|--enable|--disable|--init|--check|--dry-run` is required.

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
		t.Fatalf("expected non-zero for bare --interceptor")
	}
	assertEmptyStdout(t, resp.Stdout)
	// Soft match: action / usage / interceptor messaging
	assert.Output(t, resp.Stderr, `<contains>
interceptor
</contains>`)
}
```
