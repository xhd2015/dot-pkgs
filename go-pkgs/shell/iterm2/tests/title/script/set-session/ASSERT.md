## Expected

- Non-empty script.
- Contains session UUID from `SessionID`.
- Mentions setting a session name (includes `name` and `session`).
- Embeds the title `Hello Tab`.

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
	if resp.Script == "" {
		t.Fatal("expected non-empty BuildSetTitleScript output")
	}
	if !scriptHasUUID(resp.Script, req.SessionID) {
		t.Fatalf("missing session UUID in script:\n%s", resp.Script)
	}
	lower := strings.ToLower(resp.Script)
	if !strings.Contains(lower, "session") {
		t.Fatalf("session target script should mention session:\n%s", resp.Script)
	}
	if !strings.Contains(lower, "name") {
		t.Fatalf("script should set name:\n%s", resp.Script)
	}
	if !strings.Contains(resp.Script, "Hello Tab") {
		t.Fatalf("script should embed title:\n%s", resp.Script)
	}
}
```
