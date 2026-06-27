## Expected

Default flex mode: when upstream is dead at startup and later becomes available,
the proxy must detect it and switch to upstream.

- "falling back to direct" appears once at startup
- "upstream proxy available, switching" appears after upstream starts listening
- "falling back to direct" comes before "upstream proxy available, switching"

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := resp.Output

	if !strings.Contains(output, "falling back to direct") {
		t.Fatalf("expected initial 'falling back to direct' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "upstream proxy available, switching") {
		t.Fatalf("expected 'upstream proxy available, switching' after upstream comes up, got:\n%s", output)
	}

	idxFallback := strings.Index(output, "falling back to direct")
	idxSwitch := strings.Index(output, "upstream proxy available, switching")
	if idxFallback == -1 || idxSwitch == -1 || idxFallback >= idxSwitch {
		t.Fatalf("expected 'falling back to direct' before 'upstream proxy available, switching', got:\n%s", output)
	}
}
```