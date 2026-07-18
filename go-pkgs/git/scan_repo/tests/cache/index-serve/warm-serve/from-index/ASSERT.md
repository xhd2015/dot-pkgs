## Expected

- Scan succeeds with no RootErrors.
- Exactly one repo: `known-repo` (main), Path/GitDir match fixture.
- `resp.IndexOK` is true; index lists `known-repo` abs path under universe home.
- Proves warm path still returns the cold-seeded repo when serving via index
  (P2); RED until Scan seeds/loads `repos.json` on the warm path.

## Errors

- `err` is nil.

## Side Effects

- Index remains loadable after warm Scan (not wiped by warm serve).

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

	known := req.KnownPath
	if known == "" {
		known = absPath(t, filepath.Join(req.Roots[0], "known-repo"))
	}
	wantGit := absPath(t, filepath.Join(known, ".git"))

	if len(resp.Repos) != 1 {
		t.Fatalf("expected 1 warm-served repo, got %d: %v", len(resp.Repos), resp.Repos)
	}
	r := resp.Repos[0]
	if r.Path != known {
		t.Fatalf("Path = %q, want %q", r.Path, known)
	}
	if r.RepoType != scan_repo.RepoTypeMain {
		t.Fatalf("RepoType = %v, want main", r.RepoType)
	}
	if r.GitDir != wantGit {
		t.Fatalf("GitDir = %q, want %q", r.GitDir, wantGit)
	}

	if !resp.IndexOK {
		t.Fatal("expected IndexOK after cold seed + warm; Scan must seed home/repos.json")
	}
	if resp.Index.Universe != "home" {
		t.Fatalf("Index.Universe = %q, want home", resp.Index.Universe)
	}
	paths := indexPaths(resp.Index)
	if _, ok := paths[known]; !ok {
		t.Fatalf("index missing known-repo %q; entries=%v", known, resp.Index.Repos)
	}
}
```
