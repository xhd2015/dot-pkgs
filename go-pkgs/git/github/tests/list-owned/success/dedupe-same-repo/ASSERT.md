## Expected

- `len(resp.Repos)` is 1.
- Single repo `alice/shared` with Description `from alice` (first owner query wins).

## Side Effects

- Mock `gh` invoked for both owners.

## Errors

- `err` is nil.

## Exit Code

- N/A (library call).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Repos) != 1 {
		t.Fatalf("expected 1 repo after dedupe, got %d: %+v", len(resp.Repos), resp.Repos)
	}
	repo := resp.Repos[0]
	if repo.FullName != "alice/shared" {
		t.Fatalf("expected alice/shared, got %q", repo.FullName)
	}
	if repo.Description != "from alice" {
		t.Fatalf("expected first occurrence description %q, got %q", "from alice", repo.Description)
	}
}```