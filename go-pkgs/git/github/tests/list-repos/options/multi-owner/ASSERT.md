## Expected

- `err` is nil.
- `len(resp.Results)` is 2, sorted: `alice/zebra`, `bob/apple`.
- Every result has `matched_by: ["owned"]`.
- Captured gh argv includes both `repo list alice` and `repo list bob`.

## Side Effects

- Mock `gh api user` and two `repo list` calls.

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
	if resp.Results[0].FullName != "alice/zebra" || resp.Results[1].FullName != "bob/apple" {
		t.Fatalf("unexpected merge order: %+v", resp.Results)
	}
	for _, r := range resp.Results {
		assertMatchedBy(t, matchedByStrings(r.MatchedBy), []string{"owned"})
	}
	argv := resp.GhArgv
	if argv == "" {
		argv = readGhArgv(t, req.GhBin)
	}
	if !strings.Contains(argv, "repo list alice") || !strings.Contains(argv, "repo list bob") {
		t.Fatalf("expected repo list for both owners, argv=%q", argv)
	}
}
```