## Expected

- Script embeds tab display names `Alpha` and `Beta`.
- Script sets a session name (contains `set name` and references a session).

## Exit Code

- N/A (build-tab-set-script phase)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	s := resp.Script
	if s == "" {
		t.Fatal("expected non-empty script")
	}
	if !strings.Contains(s, "Alpha") {
		t.Fatalf("missing session name Alpha; script:\n%s", s)
	}
	if !strings.Contains(s, "Beta") {
		t.Fatalf("missing session name Beta; script:\n%s", s)
	}
	lower := strings.ToLower(s)
	if !strings.Contains(lower, "set name") {
		t.Fatalf("script should set name for sessions; script:\n%s", s)
	}
	if !strings.Contains(lower, "session") {
		t.Fatalf("session name target should mention session; script:\n%s", s)
	}
}
```
