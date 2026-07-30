## Expected

- `IsClean` returns a non-nil error (`IsCleanNil=false`); message mentions uncommitted changes
- `IsCleanWrk` is false (untracked counts as dirty under wrk taxonomy)
- Run itself returns no error (clean dirty is response data, not Run failure)

## Side Effects

- Untracked file left in temp fixture only

## Errors

- Run error must be nil; dirty is expressed via `IsClean` / `IsCleanWrk` fields

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatalf("unexpected Run error: %v", err)
	}
	if resp.IsCleanNil {
		t.Fatal("IsClean = nil, want error for untracked-only porcelain dirt")
	}
	if resp.IsCleanErr == "" {
		t.Fatal("IsCleanErr empty, want uncommitted-changes message")
	}
	if !strings.Contains(resp.IsCleanErr, "uncommitted changes") {
		t.Fatalf("IsCleanErr = %q, want substring %q", resp.IsCleanErr, "uncommitted changes")
	}
	if resp.IsCleanWrk {
		t.Fatal("IsCleanWrk = true, want false for untracked-only dirt")
	}
}
```
