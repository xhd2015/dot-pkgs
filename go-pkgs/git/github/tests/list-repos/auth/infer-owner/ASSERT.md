## Expected

- `err` is nil.
- `resp.Login` is `alice`.
- `len(resp.Results)` is 2, sorted by `FullName`: `alice/alpha`, `alice/beta`.
- Each result has `matched_by: ["owned"]`.
- Captured gh argv includes `repo list alice`.

## Side Effects

- Mock `gh api user` called; `gh repo list alice` called.

## Errors

- None.

## Exit Code

- N/A (library call).

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Login != "alice" {
		t.Fatalf("expected login alice, got %q", resp.Login)
	}
	if len(resp.Results) != 2 {
		t.Fatalf("expected 2 results, got %d: %+v", len(resp.Results), resp.Results)
	}
	names := []string{resp.Results[0].FullName, resp.Results[1].FullName}
	assertSortedFullNames(t, names)
	if resp.Results[0].FullName != "alice/alpha" || resp.Results[1].FullName != "alice/beta" {
		t.Fatalf("unexpected order: %+v", resp.Results)
	}
	for _, r := range resp.Results {
		assertMatchedBy(t, matchedByStrings(r.MatchedBy), []string{"owned"})
	}
	argv := resp.GhArgv
	if argv == "" {
		argv = readGhArgv(t, req.GhBin)
	}
	if !strings.Contains(argv, "repo list alice") {
		t.Fatalf("expected repo list alice in argv, got %q", argv)
	}
}
```