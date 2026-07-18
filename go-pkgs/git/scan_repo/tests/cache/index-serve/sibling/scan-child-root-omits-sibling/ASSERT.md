## Expected

- Scan succeeds with no RootErrors.
- Every `Result.Repos` path is under abs scan root `A` (`KnownPath`).
- Neighboring sibling checkout `B` (`SiblingPath`) **must not** appear.
- At least `A` itself is present as `RepoTypeMain` (indexed live root).
- `req.Refresh` is false — under-root filter is warm+sibling, not force-cold.
- Classic TDD: RED while sibling probe merges `B` into Result without
  `pathIsUnderRoot(absRoot, path)` drop.

## Errors

- `err` is nil.

## Side Effects

- Save-back may update home index for universe under A; not asserted here —
  Result path set is the contract.

```go
import (
	"path/filepath"
	"strings"
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
	if req.Refresh {
		t.Fatal("Refresh must be false; under-root filter is not force-cold")
	}

	a := req.KnownPath
	b := req.SiblingPath
	if a == "" || b == "" {
		t.Fatal("Setup must set KnownPath (A) and SiblingPath (B)")
	}
	if len(req.Roots) != 1 || absPath(t, req.Roots[0]) != a {
		t.Fatalf("Roots must be single child A %q, got %v", a, req.Roots)
	}

	// Contract: no path outside absRoot=A (sibling B is the critical leak).
	got := resultPaths(resp.Repos)
	if _, ok := got[b]; ok {
		t.Fatalf("Result must omit sibling B %q outside scan root A %q; repos=%v",
			b, a, resp.Repos)
	}
	for _, r := range resp.Repos {
		if r.Path != a && !strings.HasPrefix(r.Path, a+string(filepath.Separator)) {
			t.Fatalf("repo path %q not under scan root %q", r.Path, a)
		}
	}
	if _, ok := got[a]; !ok {
		t.Fatalf("Result missing cold-indexed A %q; repos=%v", a, resp.Repos)
	}
	for i, r := range resp.Repos {
		if r.Path == a && r.RepoType != scan_repo.RepoTypeMain {
			t.Fatalf("A repos[%d].RepoType = %v, want main", i, r.RepoType)
		}
	}
}
```
