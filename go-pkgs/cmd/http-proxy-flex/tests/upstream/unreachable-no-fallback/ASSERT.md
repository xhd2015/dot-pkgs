## Expected

- Exit code is 0 (server starts despite unreachable upstream)
- Logs contain "falling back to direct" (warning on bootstrap)
- Logs contain "listening on" (server started)

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\noutput:\n%s", resp.ExitCode, resp.Output)
	}
	output := resp.Output
	if !strings.Contains(output, "falling back to direct") {
		t.Fatalf("expected 'falling back to direct' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "listening on") {
		t.Fatalf("expected 'listening on' in output, got:\n%s", output)
	}
}
```
