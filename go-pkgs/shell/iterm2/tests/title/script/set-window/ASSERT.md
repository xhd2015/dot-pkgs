## Expected

- Non-empty script with UUID.
- Mentions window and name; embeds `Hello Window`.

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
	if !strings.Contains(lower, "window") {
		t.Fatalf("window target script should mention window:\n%s", resp.Script)
	}
	if !strings.Contains(lower, "name") {
		t.Fatalf("script should set name:\n%s", resp.Script)
	}
	if !strings.Contains(resp.Script, "Hello Window") {
		t.Fatalf("script should embed title:\n%s", resp.Script)
	}
}
```
