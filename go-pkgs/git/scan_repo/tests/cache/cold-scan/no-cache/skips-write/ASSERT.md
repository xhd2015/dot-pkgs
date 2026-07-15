## Expected

- Scan succeeds with exactly one discovered main repo (`my-repo`).
- No `entry.json` files exist under `CacheRoot` (mirror absent or empty).
- `LoadCacheEntry` for the repo path returns `ok=false`.

## Errors

- `err` is nil.

## Side Effects

- CacheRoot remains free of mirror entry files despite successful Scan.

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
	if len(resp.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(resp.Repos))
	}
	repoPath := absPath(t, filepath.Join(req.Roots[0], "my-repo"))
	if resp.Repos[0].Path != repoPath {
		t.Fatalf("Path = %q, want %q", resp.Repos[0].Path, repoPath)
	}

	entry, ok, loadErr := scan_repo.LoadCacheEntry(req.CacheRoot, repoPath)
	if loadErr != nil {
		t.Fatalf("LoadCacheEntry: %v", loadErr)
	}
	if ok {
		t.Fatalf("expected no cache entry under NoCache, got %+v", entry)
	}

	// Walk CacheRoot for any entry.json — none allowed when NoCache.
	var found []string
	_ = filepath.WalkDir(req.CacheRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() && d.Name() == "entry.json" {
			found = append(found, path)
		}
		return nil
	})
	if len(found) != 0 {
		t.Fatalf("NoCache=true but found entry.json files: %v", found)
	}
}
```
