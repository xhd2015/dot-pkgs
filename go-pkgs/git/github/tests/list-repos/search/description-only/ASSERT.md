## Expected

- `err` is nil.
- `len(resp.Results)` is 2, sorted: `alice/gadget`, `alice/widget-lib`.
- Every result has `matched_by: ["description"]` only.
- Captured gh argv includes `search repos` and keyword `widget`.

## Side Effects

- Mock `gh api user` and `gh search repos` invoked; `repo list` not used.

## Errors

- None.

## Exit Code

- N/A (library call).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Results) != 2 {
		t.Fatalf("expected 2 results, got %d: %+v", len(resp.Results), resp.Results)
	}
	names := []string{resp.Results[0].FullName, resp.Results[1].FullName}
	assertSortedFullNames(t, names)
	if resp.Results[0].FullName != "alice/gadget" || resp.Results[1].FullName != "alice/widget-lib" {
		t.Fatalf("unexpected repos: %+v", resp.Results)
	}
	for _, r := range resp.Results {
		assertMatchedBy(t, matchedByStrings(r.MatchedBy), []string{"description"})
	}
	argv := resp.GhArgv
	if argv == "" {
		argv = readGhArgv(t, req.GhBin)
	}
	if !strings.Contains(argv, "search repos") {
		t.Fatalf("expected search repos in argv, got %q", argv)
	}
	if !strings.Contains(argv, "widget") {
		t.Fatalf("expected search keyword widget in argv, got %q", argv)
	}
}
```