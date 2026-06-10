## Expected

- "falling back to direct" appears at least twice in the output (initial + after upstream goes down)
- "upstream proxy available, switching" appears at least once
- The first "falling back to direct" comes before "upstream proxy available, switching"
- The second "falling back to direct" comes after "upstream proxy available, switching"
- "listening on" appears (server started successfully)

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := resp.Output

	countFallback := strings.Count(output, "falling back to direct")
	if countFallback < 2 {
		t.Fatalf("expected 'falling back to direct' at least twice, got %d occurrence(s)\noutput:\n%s", countFallback, output)
	}

	if !strings.Contains(output, "upstream proxy available, switching") {
		t.Fatalf("expected 'upstream proxy available, switching' in output, got:\n%s", output)
	}

	if !strings.Contains(output, "listening on") {
		t.Fatalf("expected 'listening on' in output, got:\n%s", output)
	}

	// Verify order: fallback-1 < switching < fallback-2
	idxFallback1 := strings.Index(output, "falling back to direct")
	idxSwitch := strings.Index(output, "upstream proxy available, switching")
	idxFallback2 := strings.LastIndex(output, "falling back to direct")

	if idxFallback1 == -1 || idxSwitch == -1 || idxFallback2 == -1 {
		t.Fatalf("missing expected log lines\noutput:\n%s", output)
	}

	if idxFallback1 >= idxSwitch {
		t.Fatalf("expected first 'falling back to direct' before 'upstream proxy available, switching'\noutput:\n%s", output)
	}

	if idxSwitch >= idxFallback2 {
		t.Fatalf("expected 'upstream proxy available, switching' before second 'falling back to direct'\noutput:\n%s", output)
	}

	if resp.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", resp.ExitCode)
	}
}
```
