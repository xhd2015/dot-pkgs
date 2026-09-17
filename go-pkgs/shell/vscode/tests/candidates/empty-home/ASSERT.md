## Expected

- Exactly two candidates, both under `/Applications`.
- No `filepath.Join("", "Applications", ...)` style relative path leaks in.

## Side Effects

- None.

## Errors

- None.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	if got := len(resp.Candidates); got != 2 {
		t.Fatalf("candidates = %#v, want 2 system paths", resp.Candidates)
	}
	for _, p := range resp.Candidates {
		if !hasPrefix([]string{p}, "/Applications/") {
			t.Fatalf("candidates = %#v, want only absolute system paths", resp.Candidates)
		}
	}
}
```
