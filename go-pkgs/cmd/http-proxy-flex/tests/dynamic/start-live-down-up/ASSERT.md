## Expected

- "using upstream proxy" appears at least once (initial detection at startup)
- "falling back to direct" appears at least once (after upstream goes down)
- "upstream proxy available, switching" appears at least once (after upstream comes back)
- The first "using upstream proxy" comes before "falling back to direct"
- "falling back to direct" comes before "upstream proxy available, switching"
- "listening on" appears (server started successfully)
- Health check loop detects both down and up transitions starting from a live upstream

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

	if !strings.Contains(output, "using upstream proxy") {
		t.Fatalf("expected 'using upstream proxy' in output, got:\n%s", output)
	}

	if !strings.Contains(output, "falling back to direct") {
		t.Fatalf("expected 'falling back to direct' in output, got:\n%s", output)
	}

	if !strings.Contains(output, "upstream proxy available, switching") {
		t.Fatalf("expected 'upstream proxy available, switching' in output, got:\n%s", output)
	}

	if !strings.Contains(output, "listening on") {
		t.Fatalf("expected 'listening on' in output, got:\n%s", output)
	}

	// Verify order: using-upstream < fallback < available-switching
	idxUsing := strings.Index(output, "using upstream proxy")
	idxFallback := strings.Index(output, "falling back to direct")
	idxAvailable := strings.LastIndex(output, "upstream proxy available, switching")

	if idxUsing == -1 || idxFallback == -1 || idxAvailable == -1 {
		t.Fatalf("missing expected log lines\noutput:\n%s", output)
	}

	if idxUsing >= idxFallback {
		t.Fatalf("expected 'using upstream proxy' before 'falling back to direct'\noutput:\n%s", output)
	}

	if idxFallback >= idxAvailable {
		t.Fatalf("expected 'falling back to direct' before 'upstream proxy available, switching'\noutput:\n%s", output)
	}

	if resp.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", resp.ExitCode)
	}
}
```
