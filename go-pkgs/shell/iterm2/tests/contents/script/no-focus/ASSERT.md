## Expected

- Script contains `contents of aSession`
- Script contains the session UUID
- Script does not activate or select
- Path-bound tell when AppPath is set

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
	if s == "" {
		t.Fatal("expected non-empty script")
	}
	if !strings.Contains(s, "contents of aSession") {
		t.Fatalf("expected contents of aSession:\n%s", s)
	}
	if !strings.Contains(s, "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee") {
		t.Fatalf("missing UUID:\n%s", s)
	}
	lower := strings.ToLower(s)
	if strings.Contains(lower, "activate") {
		t.Fatalf("must not activate:\n%s", s)
	}
	if strings.Contains(lower, "select ") || strings.Contains(lower, "select\n") {
		t.Fatalf("must not select:\n%s", s)
	}
}
```
