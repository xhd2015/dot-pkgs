# Scenario

## Expected

- `EnsureInstalled` succeeds with no error.
- No `visudo` or `install` runner calls.

## Side Effects

- None.

## Errors

- None.

## Exit Code

- N/A.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if hasRunnerCall(resp.RunnerCalls, "sudo", "visudo") {
		t.Fatal("visudo must not run when already installed")
	}
	if hasRunnerCall(resp.RunnerCalls, "sudo", "install") {
		t.Fatal("install must not run when already installed")
	}
}
```