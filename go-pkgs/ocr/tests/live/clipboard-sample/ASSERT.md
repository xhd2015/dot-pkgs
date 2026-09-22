---
label: e2e
---

## Expected

- `err` is nil.
- `resp.Engine` is `vision`.
- `resp.Text` contains `How can I help` and `hihelo` (case-insensitive ok for help phrase).

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Engine != "vision" {
		t.Fatalf("Engine = %q", resp.Engine)
	}
	lower := strings.ToLower(resp.Text)
	if !strings.Contains(lower, "how can i help") {
		t.Fatalf("Text missing help phrase: %q", resp.Text)
	}
	if !strings.Contains(lower, "hihelo") {
		t.Fatalf("Text missing hihelo: %q", resp.Text)
	}
}
```
