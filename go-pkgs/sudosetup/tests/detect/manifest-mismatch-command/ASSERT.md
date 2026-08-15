## Expected

- `Status.Installed` is false.
- `Status.InstallDetail` mentions command mismatch.

## Side Effects

- None.

## Errors

- None from `Run`.

## Exit Code

- Success.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertEqual(t, "Installed", resp.Status.Installed, false)
	assertContains(t, resp.Status.InstallDetail, "command")
}
```