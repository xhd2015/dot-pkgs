## Expected

- Scan succeeds with no RootErrors.
- Result includes `a-known`, `b-known`, and `a-new` (oldest unit refreshed).
- `b-new` is **not** listed (newer unit not reached within tiny budget).

## Errors

- `err` is nil.

## Side Effects

- Proves oldest-first unit selection under a budget that fits one unit only.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(resp.RootErrors) != 0 {
		t.Fatalf("expected no RootErrors, got %v", resp.RootErrors)
	}

	aKnown := absPath(t, filepath.Join(req.Roots[0], "unit-a", "a-known"))
	bKnown := absPath(t, filepath.Join(req.Roots[0], "unit-b", "b-known"))
	aNew := absPath(t, filepath.Join(req.Roots[0], "unit-a", "a-new"))
	bNew := absPath(t, filepath.Join(req.Roots[0], "unit-b", "b-new"))

	got := map[string]bool{}
	for _, r := range resp.Repos {
		got[r.Path] = true
	}
	if got[bNew] {
		t.Fatalf("listed b-new %q; want omit (unit-b must not refresh under tiny budget)", bNew)
	}
	if !got[aKnown] {
		t.Fatalf("missing a-known %q in Result", aKnown)
	}
	if !got[bKnown] {
		t.Fatalf("missing b-known %q in Result", bKnown)
	}
	if !got[aNew] {
		t.Fatalf("missing a-new %q; oldest unit-a must be refreshed first", aNew)
	}
	if len(resp.Repos) != 3 {
		t.Fatalf("expected 3 repos (a-known, a-new, b-known), got %d: %v", len(resp.Repos), pathsOf(resp.Repos))
	}
}

func pathsOf(repos []scan_repo.Repo) []string {
	out := make([]string, len(repos))
	for i, r := range repos {
		out[i] = r.Path
	}
	return out
}
```
