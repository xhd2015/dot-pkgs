## Expected

- `Attach` returns a non-nil error when Domain is empty.
- Error text should indicate domain is required (substring match is enough).
- Caller is not required to have created managed `state.json` for this failure.

## Side Effects

- None required. Best-effort: if a managed dir somehow exists, that is not asserted.

## Errors

- Non-nil error from `Attach` / `Run`.

## Exit Code

- Non-zero API error (`err != nil` from `Run`)

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/cloudflare"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("Attach(empty Domain) err=nil; want error")
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "domain") {
		t.Fatalf("error %q should mention domain", err.Error())
	}
	// No requirement that state.json exists after validation failure.
	if req.ConfigDir != "" && req.TunnelName != "" {
		if dir, derr := cloudflare.ManagedTunnelDir(req.ConfigDir, req.TunnelName); derr == nil {
			if _, serr := os.Stat(filepath.Join(dir, "state.json")); serr == nil {
				t.Logf("note: state.json exists after validation failure (not required to be absent)")
			}
		}
	}
}
```
