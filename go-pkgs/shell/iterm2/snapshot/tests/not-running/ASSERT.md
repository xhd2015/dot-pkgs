## Expected

- Capture returns an error whose message includes `iTerm2 is not running`.
- Snapshot is nil (or unused); no successful inventory.

## Errors

- `err != nil` with substring `iTerm2 is not running`.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	if err == nil {
		t.Fatal("expected error when iTerm2 is not running")
	}
	if !strings.Contains(err.Error(), "iTerm2 is not running") {
		t.Fatalf("error %q does not contain %q", err.Error(), "iTerm2 is not running")
	}
	if resp != nil && resp.Snap != nil {
		t.Fatal("expected nil snapshot on not-running error")
	}
}
```
