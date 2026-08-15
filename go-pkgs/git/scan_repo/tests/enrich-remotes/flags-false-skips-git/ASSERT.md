## Expected

- One fake repo discovered.
- `Remotes` is empty — no git enrichment ran.

## Errors

- `err` is nil.

## Side Effects

- No `git` subprocess required; fake `.git` directory suffices.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(resp.Repos))
	}
	if len(resp.Repos[0].Remotes) != 0 {
		t.Fatalf("expected empty Remotes when ListRemotes=false, got %v", resp.Repos[0].Remotes)
	}
}
```