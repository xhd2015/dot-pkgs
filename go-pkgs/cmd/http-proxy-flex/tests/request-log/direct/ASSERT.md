## Expected

- The captured log contains a line with "via direct"

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.Output, "via direct") {
		t.Fatalf("expected 'via direct' in output, got:\n%s", resp.Output)
	}
}
```
