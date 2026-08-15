## Expected

- `len(resp.Repos)` is 2.
- Full names `alice/zebra` and `bob/apple`, sorted ascending (`alice/zebra` first).

## Side Effects

- Mock `gh` invoked for both owners.

## Errors

- `err` is nil.

## Exit Code

- N/A (library call).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Repos) != 2 {
		t.Fatalf("expected 2 repos, got %d: %+v", len(resp.Repos), resp.Repos)
	}
	names := []string{resp.Repos[0].FullName, resp.Repos[1].FullName}
	assertSortedFullNames(t, names)
	if resp.Repos[0].FullName != "alice/zebra" {
		t.Fatalf("expected alice/zebra first, got %q", resp.Repos[0].FullName)
	}
	if resp.Repos[1].FullName != "bob/apple" {
		t.Fatalf("expected bob/apple second, got %q", resp.Repos[1].FullName)
	}
}```