## Expected

- `err` is nil.
- `resp.Repos` is empty.
- Exactly one `RootError` indicating the root is not a directory.

## Errors

- Scan returns fatal error instead of recording RootError.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("expected nil scan error, got %v", err)
	}
	if len(resp.Repos) != 0 {
		t.Fatalf("expected 0 repos, got %d", len(resp.Repos))
	}
	if len(resp.RootErrors) != 1 {
		t.Fatalf("expected 1 root error, got %d: %v", len(resp.RootErrors), resp.RootErrors)
	}
	fileRoot := req.Roots[0]
	re := resp.RootErrors[0]
	if re.Root != fileRoot {
		t.Fatalf("RootError.Root = %q, want %q", re.Root, fileRoot)
	}
	msg := strings.ToLower(re.Error)
	if !strings.Contains(msg, "directory") {
		t.Fatalf("RootError.Error should mention directory, got %q", re.Error)
	}
}
```