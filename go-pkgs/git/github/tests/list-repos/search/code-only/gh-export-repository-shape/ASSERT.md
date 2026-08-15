## Expected

- `err` is nil.
- `len(resp.Results)` is 2 after dedupe, sorted: `xhd2015/dot-pkgs`, `xhd2015/lifelog`.
- Owner parsed from `nameWithOwner` (not `owner.login`).
- Every result has `matched_by: ["code"]` only.

## Side Effects

- Mock `gh api user` and `gh search code` invoked.

## Errors

- Must **not** return `missing repository owner`.

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
		if strings.Contains(err.Error(), "missing repository owner") {
			t.Fatalf("bug reproduced: parseSearchCode rejects gh export shape: %v", err)
		}
		t.Fatal(err)
	}
	if len(resp.Results) != 2 {
		t.Fatalf("expected 2 deduped results, got %d: %+v", len(resp.Results), resp.Results)
	}
	names := []string{resp.Results[0].FullName, resp.Results[1].FullName}
	assertSortedFullNames(t, names)
	if resp.Results[0].FullName != "xhd2015/dot-pkgs" || resp.Results[1].FullName != "xhd2015/lifelog" {
		t.Fatalf("unexpected repos: %+v", resp.Results)
	}
	if resp.Results[0].Owner != "xhd2015" || resp.Results[1].Owner != "xhd2015" {
		t.Fatalf("expected owner xhd2015 from nameWithOwner, got: %+v", resp.Results)
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
	if !strings.Contains(argv, "doctest") {
		t.Fatalf("expected search keyword doctest in argv, got %q", argv)
	}
}
```