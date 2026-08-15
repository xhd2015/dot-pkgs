## Expected

- `Remove` succeeds without error.
- Runner records `sudo -k` only (no `rm`).
- `Installed` remains false.

## Side Effects

- Timestamp cache flush via `sudo -k`.

## Errors

- None.

## Exit Code

- Success.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertEqual(t, "Installed", resp.Installed, false)
	if !hasRunnerCall(resp.RunnerCalls, "sudo", "-k") {
		t.Fatal("expected sudo -k call")
	}
	if hasRunnerCall(resp.RunnerCalls, "sudo", "rm") {
		t.Fatal("sudo rm must not run when nothing installed")
	}
}
```