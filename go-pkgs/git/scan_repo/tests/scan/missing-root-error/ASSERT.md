## Expected

- `err` is nil.
- `resp.Repos` is empty.
- Exactly one `RootError` for the missing root path.

## Errors

- Scan returns fatal error instead of recording RootError.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("expected nil scan error, got %v", err)
	}
	if len(resp.Repos) != 0 {
		t.Fatalf("expected 0 repos, got %d", len(resp.Repos))
	}
	if len(resp.RootErrors) != 1 {
		t.Fatalf("expected 1 root error, got %d: %v", len(resp.RootErrors), resp.RootErrors)
	}
	missing := req.Roots[0]
	re := resp.RootErrors[0]
	if re.Root != missing {
		t.Fatalf("RootError.Root = %q, want %q", re.Root, missing)
	}
	if !strings.Contains(re.Error, missing) {
		t.Fatalf("RootError.Error should contain root path %q, got %q", missing, re.Error)
	}
}
```