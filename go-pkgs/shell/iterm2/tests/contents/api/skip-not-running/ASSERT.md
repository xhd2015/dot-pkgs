## Expected

- One Exec (system only)
- App is system
- Home path is not in any script

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExecN != 1 {
		t.Fatalf("exec n = %d want 1", resp.ExecN)
	}
	if resp.App != iterm2.CanonicalITermAppSystem {
		t.Fatalf("app = %q", resp.App)
	}
	for _, s := range resp.Scripts {
		if strings.Contains(s, "/Users/me/Applications/iTerm.app") {
			t.Fatalf("must not tell not-running home:\n%s", s)
		}
	}
}
```
