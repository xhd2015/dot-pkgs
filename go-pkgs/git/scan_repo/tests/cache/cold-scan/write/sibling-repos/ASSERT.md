## Expected

- Exactly two repos in `resp.Repos`, path-sorted: `alpha` then `zebra`.
- Both are `RepoTypeMain`.
- Each has a mirror entry with `is_repo=true`, `repo_type="main"`, matching `git_dir`.

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
	for i, r := range resp.Repos {
		if r.RepoType != scan_repo.RepoTypeMain {
			t.Fatalf("repos[%d].RepoType = %v, want main", i, r.RepoType)
		}
		entry, ok, loadErr := scan_repo.LoadCacheEntry(req.CacheRoot, r.Path)
		if loadErr != nil {
			t.Fatalf("LoadCacheEntry(%s): %v", r.Path, loadErr)
		}
		if !ok {
			t.Fatalf("expected cache entry for %s", r.Path)
		}
		if !entry.IsRepo {
			t.Fatalf("%s: IsRepo=false, want true", r.Path)
		}
		if entry.RepoType != string(scan_repo.RepoTypeMain) {
			t.Fatalf("%s: RepoType=%q, want main", r.Path, entry.RepoType)
		}
		if entry.GitDir != r.GitDir {
			t.Fatalf("%s: cache GitDir=%q, want %q", r.Path, entry.GitDir, r.GitDir)
		}
		if !entry.ScanComplete {
			t.Fatalf("%s: ScanComplete=false", r.Path)
		}
	}
}
```
