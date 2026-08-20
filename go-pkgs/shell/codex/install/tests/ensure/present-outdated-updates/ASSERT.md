## Expected

- `err == nil`.
- `resp.Action == "update"`.
- `resp.ShellCalls` is exactly `["<BinPath> update"]` (path-qualified).
- `resp.FetchLatestCalls >= 1`.
- `resp.ResultNeedsUpdate == true`.
- `resp.LocalVersion` parses to `0.1.0` (or equals that form).
- `resp.LatestVersion == "0.2.0"`.

## Errors

- None.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertNoError(t, err)
	assertEqual(t, "Action", resp.Action, "update")
	bin := req.BinPath
	if bin == "" {
		bin = filepath.Join(req.WorkDir, "bin", "codex")
	}
	assertShellCalls(t, resp.ShellCalls, bin+" update")
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
