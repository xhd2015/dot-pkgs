## Expected

- One repo discovered.
- `Remotes` is empty (nil or zero-length).

## Errors

- `err` is nil.

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
		t.Fatalf("expected empty Remotes, got %v", resp.Repos[0].Remotes)
	}
}
```