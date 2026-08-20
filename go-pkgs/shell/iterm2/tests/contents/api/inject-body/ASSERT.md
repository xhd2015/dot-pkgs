## Expected

- Contents is the injected body
- App is the home canonical tag
- Only one Exec (home hit, system not queried)

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Contents != "❯ hello pane" {
		t.Fatalf("contents = %q", resp.Contents)
	}
	if resp.App != iterm2.CanonicalITermAppHome {
		t.Fatalf("app = %q", resp.App)
	}
	if resp.SessionID != defaultContentsSessionID {
		t.Fatalf("session_id = %q", resp.SessionID)
	}
	if resp.ExecN != 1 {
		t.Fatalf("exec n = %d want 1", resp.ExecN)
	}
}
```
