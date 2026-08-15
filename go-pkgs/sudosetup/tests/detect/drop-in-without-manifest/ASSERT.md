## Expected

- `Status.Installed` is false.
- `Status.InstallDetail` mentions missing manifest or orphaned drop-in.

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
	if !detailContains(resp.Status.InstallDetail, "manifest") && !detailContains(resp.Status.InstallDetail, "orphan") {
		t.Fatalf("InstallDetail should mention manifest or orphan, got %q", resp.Status.InstallDetail)
	}
}
```