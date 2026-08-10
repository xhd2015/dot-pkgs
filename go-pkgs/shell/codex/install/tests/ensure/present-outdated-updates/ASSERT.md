## Expected

- `err == nil`.
- `resp.Action == "update"`.
- `resp.ShellCalls` is exactly `[install.UpdateCmd]`.
- `resp.FetchLatestCalls >= 1`.
- `resp.ResultNeedsUpdate == true`.
- `resp.LocalVersion` parses to `0.1.0` (or equals that form).
- `resp.LatestVersion == "0.2.0"`.

## Errors

- None.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/codex/install"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	assertNoError(t, err)
	assertEqual(t, "Action", resp.Action, "update")
	assertShellCalls(t, resp.ShellCalls, install.UpdateCmd)
	if resp.FetchLatestCalls < 1 {
		t.Fatalf("FetchLatestCalls = %d, want >= 1", resp.FetchLatestCalls)
	}
	assertEqual(t, "ResultNeedsUpdate", resp.ResultNeedsUpdate, true)
	assertEqual(t, "LatestVersion", resp.LatestVersion, "0.2.0")
	// LocalVersion may be parsed or raw; accept either form for 0.1.0
	if resp.LocalVersion != "0.1.0" && resp.LocalVersion != "codex-cli 0.1.0" {
		t.Fatalf("LocalVersion = %q, want 0.1.0 or codex-cli 0.1.0", resp.LocalVersion)
	}
}
```
