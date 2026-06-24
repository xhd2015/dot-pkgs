## Expected

- `err` is nil.
- `len(resp.Results)` is 2, sorted: `alice/alpha`, `alice/beta`.
- Every result has `matched_by: ["owned"]` only.
- No `search repos` or `search code` in captured gh argv.

## Side Effects

- Mock `gh api user` and `gh repo list alice` invoked.

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
	if resp.Results[0].FullName != "alice/alpha" || resp.Results[1].FullName != "alice/beta" {
		t.Fatalf("unexpected repos: %+v", resp.Results)
	}
	for _, r := range resp.Results {
		assertMatchedBy(t, matchedByStrings(r.MatchedBy), []string{"owned"})
	}
	argv := resp.GhArgv
	if argv == "" {
		argv = readGhArgv(t, req.GhBin)
	}
	if strings.Contains(argv, "search repos") || strings.Contains(argv, "search code") {
		t.Fatalf("search subcommands should not run in plain mode, argv=%q", argv)
	}
}
```