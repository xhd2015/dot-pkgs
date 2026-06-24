## Expected

- `err` is nil.
- `len(resp.Results)` is 2 after dedupe, sorted: `alice/tool-a`, `alice/tool-b`.
- Every result has `matched_by: ["code"]` only.
- Captured gh argv includes `search code` and keyword `handler`.

## Side Effects

- Mock `gh api user` and `gh search code` invoked.

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
		t.Fatalf("expected 2 deduped results, got %d: %+v", len(resp.Results), resp.Results)
	}
	names := []string{resp.Results[0].FullName, resp.Results[1].FullName}
	assertSortedFullNames(t, names)
	if resp.Results[0].FullName != "alice/tool-a" || resp.Results[1].FullName != "alice/tool-b" {
		t.Fatalf("unexpected repos: %+v", resp.Results)
	}
	for _, r := range resp.Results {
		assertMatchedBy(t, matchedByStrings(r.MatchedBy), []string{"code"})
	}
	argv := resp.GhArgv
	if argv == "" {
		argv = readGhArgv(t, req.GhBin)
	}
	if !strings.Contains(argv, "search code") {
		t.Fatalf("expected search code in argv, got %q", argv)
	}
	if !strings.Contains(argv, "handler") {
		t.Fatalf("expected search keyword handler in argv, got %q", argv)
	}
}
```