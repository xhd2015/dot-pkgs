## Expected

- `List` returns 2 entries including main and linked
- Main entry has `IsMain=true`; linked has `IsMain=false` and Branch `feature`
- `ListLinked` returns exactly 1 entry: the linked path (not main)
- No error from Run

## Side Effects

- None beyond temp fixture (cleaned by `t.Cleanup`)

## Errors

- None expected

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Entries) != 2 {
		t.Fatalf("List len = %d, want 2; paths=%v", len(resp.Entries), entryPaths(resp.Entries))
	}
	if !containsPath(t, resp.Entries, req.MainPath) {
		t.Fatalf("List missing main %q; paths=%v", req.MainPath, entryPaths(resp.Entries))
	}
	if !containsPath(t, resp.Entries, req.LinkedPath) {
		t.Fatalf("List missing linked %q; paths=%v", req.LinkedPath, entryPaths(resp.Entries))
	}
	mainE, ok := findByPath(t, resp.Entries, req.MainPath)
	if !ok || !mainE.IsMain {
		t.Fatalf("main entry IsMain=false or missing: %+v", mainE)
	}
	linkedE, ok := findByPath(t, resp.Entries, req.LinkedPath)
	if !ok {
		t.Fatal("linked entry missing from List")
	}
	if linkedE.IsMain {
		t.Fatal("linked entry IsMain=true, want false")
	}
	if linkedE.Branch != "feature" {
		t.Fatalf("linked Branch = %q, want feature", linkedE.Branch)
	}

	if len(resp.Linked) != 1 {
		t.Fatalf("ListLinked len = %d, want 1; paths=%v", len(resp.Linked), entryPaths(resp.Linked))
	}
	if !samePath(t, resp.Linked[0].Path, req.LinkedPath) {
		t.Fatalf("ListLinked path = %q, want %q", resp.Linked[0].Path, req.LinkedPath)
	}
	if resp.Linked[0].IsMain {
		t.Fatal("ListLinked entry IsMain=true, want false")
	}
	if containsPath(t, resp.Linked, req.MainPath) {
		t.Fatal("ListLinked must exclude main")
	}
}
```
