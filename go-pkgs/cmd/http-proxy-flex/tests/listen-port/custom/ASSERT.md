## Expected

- Logs contain "listening on :7829" (the custom port)
- Logs contain "upstream proxy unreachable, falling back to direct"

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := resp.Output
	if !strings.Contains(output, "listening on :7829") {
		t.Fatalf("expected 'listening on :7829' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "falling back to direct") {
		t.Fatalf("expected 'falling back to direct' in output, got:\n%s", output)
	}
}
```
