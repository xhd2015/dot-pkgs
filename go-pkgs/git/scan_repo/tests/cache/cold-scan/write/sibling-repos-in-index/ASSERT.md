## Expected

- Exactly two repos in `resp.Repos`, path-sorted: `alpha` then `zebra`.
- Both are `RepoTypeMain`.
- `home/repos.json` includes both paths with `repo_type=main` and matching `git_dir`.
- No `mirror/` under CacheRoot.

## Errors

- `err` is nil.

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
	if len(resp.Repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(resp.Repos))
	}
	alpha := absPath(t, filepath.Join(req.Roots[0], "alpha"))
	zebra := absPath(t, filepath.Join(req.Roots[0], "zebra"))
	if resp.Repos[0].Path != alpha {
		t.Fatalf("repos[0].Path = %q, want %q", resp.Repos[0].Path, alpha)
	}
	if resp.Repos[1].Path != zebra {
		t.Fatalf("repos[1].Path = %q, want %q", resp.Repos[1].Path, zebra)
	}

	idx, ok, loadErr := scan_repo.LoadRepoIndex(req.CacheRoot, scan_repo.UniverseHome)
	if loadErr != nil {
		t.Fatalf("LoadRepoIndex: %v", loadErr)
	}
	if !ok {
		t.Fatal("expected home/repos.json after cold Scan")
	}
	paths := map[string]scan_repo.RepoIndexEntry{}
	for _, e := range idx.Repos {
		paths[e.Path] = e
	}
	for _, r := range resp.Repos {
		if r.RepoType != scan_repo.RepoTypeMain {
			t.Fatalf("%s: RepoType = %v, want main", r.Path, r.RepoType)
		}
		ie, ok := paths[r.Path]
		if !ok {
			t.Fatalf("index missing %s", r.Path)
		}
		if ie.RepoType != string(scan_repo.RepoTypeMain) {
			t.Fatalf("%s index RepoType=%q", r.Path, ie.RepoType)
		}
		if ie.GitDir != r.GitDir {
			t.Fatalf("%s index GitDir=%q, want %q", r.Path, ie.GitDir, r.GitDir)
		}
	}
	assertNoMirrorDir(t, req.CacheRoot)
}
```
