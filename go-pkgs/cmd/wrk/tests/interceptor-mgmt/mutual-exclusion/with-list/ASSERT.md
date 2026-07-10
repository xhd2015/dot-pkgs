## Expected

- Non-zero exit code.
- Empty stdout.
- Stderr mentions mutual exclusion (interceptor and/or list).

## Errors

- Management mode cannot run alongside `--list`.

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
		t.Fatalf("expected non-zero for mutual exclusion, stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assertEmptyStdout(t, resp.Stdout)
	assert.Output(t, resp.Stderr, `<contains>
mutually exclusive
</contains>`)
}
```
