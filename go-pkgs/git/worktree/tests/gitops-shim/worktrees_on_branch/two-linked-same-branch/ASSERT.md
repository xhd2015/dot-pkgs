## Expected

- `WorktreesOnBranch(feature)` length 2
- Both linked paths present; main not included (still on master)
- Every returned entry has Branch `feature`
- No Run error (policy-free multi-checkout data)

## Side Effects

- None beyond temp fixture

## Errors

- None expected from Run

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Entries) != 2 {
		t.Fatalf("WorktreesOnBranch(feature) len = %d, want 2; paths=%v",
			len(resp.Entries), entryPaths(resp.Entries))
	}
	if !containsPath(t, resp.Entries, req.LinkedPath) {
		t.Fatalf("missing linked1 %q; paths=%v", req.LinkedPath, entryPaths(resp.Entries))
	}
	if !containsPath(t, resp.Entries, req.LinkedPath2) {
		t.Fatalf("missing linked2 %q; paths=%v", req.LinkedPath2, entryPaths(resp.Entries))
	}
	if containsPath(t, resp.Entries, req.MainPath) {
		t.Fatal("main must not appear on feature when still on master")
	}
	for _, e := range resp.Entries {
		if e.Branch != "feature" {
			t.Fatalf("entry %q Branch = %q, want feature", e.Path, e.Branch)
		}
	}
}
```
