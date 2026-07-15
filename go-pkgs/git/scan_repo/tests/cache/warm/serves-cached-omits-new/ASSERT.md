## Expected

- Scan succeeds with no RootErrors.
- Exactly one repo in `resp.Repos`: the previously cached `known-repo`.
- `known-repo` is `RepoTypeMain` with Path/GitDir matching the fixture.
- `brand-new-repo` is **not** listed (warm serves from cache; does not full re-walk).

## Errors

- `err` is nil.

## Side Effects

- Documents soft incompleteness: uncached repos planted after cold seed are missed
  until a cold re-scan or later budgeted refresh (P4, out of scope).

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

	knownPath := absPath(t, filepath.Join(req.Roots[0], "known-repo"))
	brandNewPath := absPath(t, filepath.Join(req.Roots[0], "brand-new-repo"))
	wantGitDir := absPath(t, filepath.Join(knownPath, ".git"))

	// Brand-new must be omitted — proves Scan did not full re-walk the tree.
	for i, r := range resp.Repos {
		if r.Path == brandNewPath {
			t.Fatalf("warm Scan listed uncached brand-new-repo at repos[%d]=%q; want omit (no full re-walk)", i, r.Path)
		}
	}

	if len(resp.Repos) != 1 {
		t.Fatalf("expected exactly 1 cached live repo, got %d: %v", len(resp.Repos), pathsOf(resp.Repos))
	}
	r := resp.Repos[0]
	if r.Path != knownPath {
		t.Fatalf("Repos[0].Path = %q, want %q", r.Path, knownPath)
	}
	if r.Name != "known-repo" {
		t.Fatalf("Name = %q, want known-repo", r.Name)
	}
	if r.RepoType != scan_repo.RepoTypeMain {
		t.Fatalf("RepoType = %v, want main", r.RepoType)
	}
	if r.GitDir != wantGitDir {
		t.Fatalf("GitDir = %q, want %q", r.GitDir, wantGitDir)
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
