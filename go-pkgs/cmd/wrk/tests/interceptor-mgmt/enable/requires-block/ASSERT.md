## Expected

- Non-zero exit code.
- Empty stdout.
- Stderr mentions missing interceptor and preferably `--init`.

## Errors

- Cannot enable without an interceptor block.

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
		t.Fatalf("expected non-zero when no interceptor block")
	}
	assertEmptyStdout(t, resp.Stdout)
	assert.Output(t, resp.Stderr, `<contains>
init
</contains>`)
	assertMgmtConfigAbsent(t, req.WrkHome)
}
```
