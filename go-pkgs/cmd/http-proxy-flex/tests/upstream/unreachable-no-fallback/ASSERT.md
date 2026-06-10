## Expected

- Exit code is non-zero (error state)
- Stderr contains "upstream proxy unreachable" or similar error message

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit code, got 0\noutput:\n%s", resp.Output)
	}
	output := strings.ToLower(resp.Output)
	if !strings.Contains(output, "unreachable") && !strings.Contains(output, "connect") && !strings.Contains(output, "refused") {
		t.Fatalf("expected error about unreachable upstream, got:\n%s", resp.Output)
	}
}
```
