## Expected

- Scan succeeds with no RootErrors.
- `Result.Repos` has exactly two mains: `alpha` and `zebra` (path-sorted).
- After Scan, `resp.IndexOK` is true and `resp.IndexPath` is
  `<CacheRoot>/home/repos.json`.
- Index envelope: `Universe == "home"`, `Version == 1`, `Base` equals abs of
  the scan root.
- Index `repos[]` includes both main abs paths (at least; extra fields optional).

## Errors

- `err` is nil.

## Side Effects

- Durable index file exists under the home universe (P2 cold seed).

```go
import (
	"os"
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

	rootAbs := absPath(t, req.Roots[0])
	alpha := absPath(t, filepath.Join(rootAbs, "alpha"))
	zebra := absPath(t, filepath.Join(rootAbs, "zebra"))

	if len(resp.Repos) != 2 {
		t.Fatalf("expected 2 Result repos, got %d", len(resp.Repos))
	}
	// Path-sorted: alpha before zebra
	if resp.Repos[0].Path != alpha || resp.Repos[1].Path != zebra {
		t.Fatalf("Result paths = [%q, %q], want [%q, %q]",
			resp.Repos[0].Path, resp.Repos[1].Path, alpha, zebra)
	}
	for i, r := range resp.Repos {
		if r.RepoType != scan_repo.RepoTypeMain {
			t.Fatalf("repos[%d].RepoType = %v, want main", i, r.RepoType)
		}
	}

	wantIndexPath := filepath.Join(req.CacheRoot, "home", "repos.json")
	if resp.IndexPath != wantIndexPath {
		t.Fatalf("IndexPath = %q, want %q", resp.IndexPath, wantIndexPath)
	}
	if !resp.IndexOK {
		t.Fatal("expected IndexOK=true after cold Scan seeds home/repos.json")
	}
	if _, statErr := os.Stat(wantIndexPath); statErr != nil {
		t.Fatalf("home/repos.json missing after cold Scan: %v", statErr)
	}

	if resp.Index.Universe != "home" {
		t.Fatalf("Index.Universe = %q, want home", resp.Index.Universe)
	}
	if resp.Index.Version != 1 {
		t.Fatalf("Index.Version = %d, want 1", resp.Index.Version)
	}
	if absPath(t, resp.Index.Base) != rootAbs {
		t.Fatalf("Index.Base = %q, want scan root %q", resp.Index.Base, rootAbs)
	}

	paths := indexPaths(resp.Index)
	if _, ok := paths[alpha]; !ok {
		t.Fatalf("index missing alpha %q; entries=%v", alpha, resp.Index.Repos)
	}
	if _, ok := paths[zebra]; !ok {
		t.Fatalf("index missing zebra %q; entries=%v", zebra, resp.Index.Repos)
	}
}
```
