## Expected

- Two repos discovered (main + worktree).
- All `Worktrees` slices are empty — no git enrichment ran.

## Errors

- `err` is nil.

## Side Effects

- No `git` subprocess required; fake `.git` fixtures suffice.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(resp.Repos))
	}
	for _, r := range resp.Repos {
		if len(r.Worktrees) != 0 {
			t.Fatalf("%s: expected empty Worktrees when ListWorktrees=false, got %v", r.Name, r.Worktrees)
		}
	}
}
```