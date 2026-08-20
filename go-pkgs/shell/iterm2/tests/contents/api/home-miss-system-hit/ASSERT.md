## Expected

- Two Exec calls
- Result app is system canonical
- Contents is system pane

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
	if resp.ExecN != 2 {
		t.Fatalf("exec n = %d want 2", resp.ExecN)
	}
	if resp.App != iterm2.CanonicalITermAppSystem {
		t.Fatalf("app = %q", resp.App)
	}
	if resp.Contents != "system pane" {
		t.Fatalf("contents = %q", resp.Contents)
	}
}
```
