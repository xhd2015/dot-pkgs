## Expected

- Non-empty get script with UUID; reads session name (mentions session + name).

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
		t.Fatal("expected non-empty BuildGetTitleScript output")
	}
	if !scriptHasUUID(resp.Script, req.SessionID) {
		t.Fatalf("missing session UUID in script:\n%s", resp.Script)
	}
	lower := strings.ToLower(resp.Script)
	if !strings.Contains(lower, "session") {
		t.Fatalf("session get script should mention session:\n%s", resp.Script)
	}
	if !strings.Contains(lower, "name") {
		t.Fatalf("script should read name:\n%s", resp.Script)
	}
}
```
