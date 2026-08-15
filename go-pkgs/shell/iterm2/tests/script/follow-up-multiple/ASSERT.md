## Expected

- `grok` appears before `codex` in the script.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	s := resp.Script
	iGrok := strings.Index(s, `write text "grok"`)
	iCodex := strings.Index(s, `write text "codex"`)
	if iGrok < 0 || iCodex < 0 {
		t.Fatalf("missing follow-ups: %q", s)
	}
	if iGrok > iCodex {
		t.Fatal("grok must precede codex in script")
	}
}
```