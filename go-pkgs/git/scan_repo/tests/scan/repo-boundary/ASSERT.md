## Expected

- Exactly one repo: `outer` only.
- Nested `inner` repo is not discovered.

## Errors

- `err` is nil.

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Repos) != 1 {
		t.Fatalf("expected 1 repo (outer only), got %d: %v", len(resp.Repos), resp.Repos)
	}
	want := absPath(t, filepath.Join(req.Roots[0], "outer"))
	if resp.Repos[0].Path != want {
		t.Fatalf("Path = %q, want %q", resp.Repos[0].Path, want)
	}
	if resp.Repos[0].Name != "outer" {
		t.Fatalf("Name = %q, want outer", resp.Repos[0].Name)
	}
}
```