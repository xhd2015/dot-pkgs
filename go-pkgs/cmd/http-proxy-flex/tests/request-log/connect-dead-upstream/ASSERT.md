## Expected

- CONNECT request logs "via direct" (upstream was never listening at startup)
- CONNECT succeeds with "200 Connection Established"

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "CONNECT " + req.ConnectTarget + " via direct"
	if !strings.Contains(resp.Output, want) {
		t.Fatalf("expected %q, got:\n%s", want, resp.Output)
	}
	if !strings.Contains(req.ConnectResponse, "200 Connection Established") {
		t.Fatalf("expected 200 Connection Established, got:\n%s", req.ConnectResponse)
	}
}
```