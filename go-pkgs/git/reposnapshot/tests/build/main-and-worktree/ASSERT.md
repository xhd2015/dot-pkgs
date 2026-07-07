## Expected

- Exactly one top-level `Node` for path `main`.
- Main node `Checkout.Branch` is `main`, `Checkout.Status` is `clean`, `Checkout.Error` empty.
- Main node has one nested worktree with path `feature-a`.
- Worktree `Checkout.Branch` is `feature/foo`, `Checkout.Status` is `clean`.
- No synthetic root-error nodes.

## Errors

- `err` is nil.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	nodes := resp.Snapshot.Nodes
	if len(nodes) != 1 {
		t.Fatalf("expected 1 top-level node, got %d: %+v", len(nodes), nodes)
	}
	main := nodes[0]
	if main.Path != "main" {
		t.Fatalf("main.Path = %q, want main", main.Path)
	}
	if main.Error != "" {
		t.Fatalf("main.Error = %q, want empty", main.Error)
	}
	if main.Checkout.Error != "" {
		t.Fatalf("main.Checkout.Error = %q, want empty", main.Checkout.Error)
	}
	if main.Checkout.Branch != "main" {
		t.Fatalf("main.Checkout.Branch = %q, want main", main.Checkout.Branch)
	}
	if main.Checkout.Status != "clean" {
		t.Fatalf("main.Checkout.Status = %q, want clean", main.Checkout.Status)
	}
	if len(main.Worktrees) != 1 {
		t.Fatalf("expected 1 nested worktree, got %d", len(main.Worktrees))
	}
	wt := main.Worktrees[0]
	if wt.Path != "feature-a" {
		t.Fatalf("worktree.Path = %q, want feature-a", wt.Path)
	}
	if wt.Checkout.Branch != "feature/foo" {
		t.Fatalf("worktree.Checkout.Branch = %q, want feature/foo", wt.Checkout.Branch)
	}
	if wt.Checkout.Status != "clean" {
		t.Fatalf("worktree.Checkout.Status = %q, want clean", wt.Checkout.Status)
	}
	if len(resp.Snapshot.RootErrors) != 0 {
		t.Fatalf("expected no RootErrors, got %+v", resp.Snapshot.RootErrors)
	}
}
```
