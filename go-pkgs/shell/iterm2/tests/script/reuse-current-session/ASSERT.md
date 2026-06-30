## Expected

- Script uses `current session of current tab of current window`.
- Script does not scan session `path` or `create tab`.
- Script cds via `write text ("cd " & quoted form of targetDir)`.

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
	if !strings.Contains(s, "current session of current tab of current window") {
		t.Fatalf("missing current session target: %q", s)
	}
	if strings.Contains(s, `variable named "path"`) {
		t.Fatal("reuse script must not scan session path")
	}
	if strings.Contains(s, "create tab with default profile") {
		t.Fatal("reuse script must not create tab")
	}
	if !strings.Contains(s, `write text ("cd " & quoted form of targetDir)`) {
		t.Fatalf("missing cd write text: %q", s)
	}
}
```