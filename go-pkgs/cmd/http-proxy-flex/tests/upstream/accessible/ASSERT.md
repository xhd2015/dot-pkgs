## Expected

- Logs contain "using upstream proxy" (upstream was reachable at startup)
- Logs contain "listening on" (server started)

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := resp.Output
	if !strings.Contains(output, "using upstream proxy") {
		t.Fatalf("expected 'using upstream proxy' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "listening on") {
		t.Fatalf("expected 'listening on' in output, got:\n%s", output)
	}
}
```
