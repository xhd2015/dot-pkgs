## Expected

- Exit code is 0 (server starts despite unreachable upstream)
- Logs contain "upstream proxy unreachable" but NOT "falling back to direct"
- Logs contain "listening on" (server started)

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\noutput:\n%s", resp.ExitCode, resp.Output)
	}
	output := resp.Output
	if strings.Contains(output, "falling back to direct") {
		t.Fatalf("did not expect 'falling back to direct' with --no-fallback-direct, got:\n%s", output)
	}
	if !strings.Contains(output, "upstream proxy unreachable") {
		t.Fatalf("expected 'upstream proxy unreachable' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "listening on") {
		t.Fatalf("expected 'listening on' in output, got:\n%s", output)
	}
}
```