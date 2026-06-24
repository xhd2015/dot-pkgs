## Expected

- `err` is nil.
- `len(resp.Results)` is 3, sorted: `alice/code-only`, `alice/desc-only`, `alice/shared`.
- `alice/shared` has `matched_by: ["description","code"]`.
- `alice/desc-only` has `matched_by: ["description"]`.
- `alice/code-only` has `matched_by: ["code"]`.
- Captured gh argv includes both `search repos` and `search code`.

## Side Effects

- Mock runs auth, description search, and code search.

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
	if len(resp.Results) != 3 {
		t.Fatalf("expected 3 union results, got %d: %+v", len(resp.Results), resp.Results)
	}
	names := make([]string, len(resp.Results))
	for i, r := range resp.Results {
		names[i] = r.FullName
	}
	assertSortedFullNames(t, names)
	if names[0] != "alice/code-only" || names[1] != "alice/desc-only" || names[2] != "alice/shared" {
		t.Fatalf("unexpected sort order: %v", names)
	}

	shared := resultByFullName(t, resp.Results, "alice/shared")
	assertMatchedBy(t, matchedByStrings(shared.MatchedBy), []string{"description", "code"})

	descOnly := resultByFullName(t, resp.Results, "alice/desc-only")
	assertMatchedBy(t, matchedByStrings(descOnly.MatchedBy), []string{"description"})

	codeOnly := resultByFullName(t, resp.Results, "alice/code-only")
	assertMatchedBy(t, matchedByStrings(codeOnly.MatchedBy), []string{"code"})

	argv := resp.GhArgv
	if argv == "" {
		argv = readGhArgv(t, req.GhBin)
	}
	if !strings.Contains(argv, "search repos") || !strings.Contains(argv, "search code") {
		t.Fatalf("expected both search subcommands in argv, got %q", argv)
	}
}
```