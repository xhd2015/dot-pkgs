## Expected

- Scan succeeds.
- `LoadCacheEntry(CacheRoot, scanRoot)` returns `ok=true`.
- Root entry: `is_repo=false`, `scan_complete=true`, `version=1`,
  `refreshed_at` non-empty, and `children` includes `"my-repo"`
  (and `"notes"` when child directories are recorded).

## Errors

- `err` is nil.

## Side Effects

- Non-repo scan root has a mirror `entry.json` (not only the repo child).

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	rootPath := absPath(t, req.Roots[0])

	entry, ok, loadErr := scan_repo.LoadCacheEntry(req.CacheRoot, rootPath)
	if loadErr != nil {
		t.Fatalf("LoadCacheEntry(root): %v", loadErr)
	}
	if !ok {
		t.Fatalf("expected cache entry for intermediate root %s", rootPath)
	}
	if entry.Version != 1 {
		t.Fatalf("Version = %d, want 1", entry.Version)
	}
	if entry.IsRepo {
		t.Fatal("IsRepo = true for scan root, want false")
	}
	if !entry.ScanComplete {
		t.Fatal("ScanComplete = false, want true")
	}
	if entry.RefreshedAt == "" {
		t.Fatal("RefreshedAt empty, want non-empty RFC3339")
	}
	if _, parseErr := time.Parse(time.RFC3339, entry.RefreshedAt); parseErr != nil {
		if _, parseErr2 := time.Parse(time.RFC3339Nano, entry.RefreshedAt); parseErr2 != nil {
			t.Fatalf("RefreshedAt %q not RFC3339: %v", entry.RefreshedAt, parseErr)
		}
	}
	if entry.Children == nil {
		t.Fatal("Children is nil, want non-nil list of child directory basenames")
	}
	hasMyRepo := false
	hasNotes := false
	for _, c := range entry.Children {
		if c == "my-repo" {
			hasMyRepo = true
		}
		if c == "notes" {
			hasNotes = true
		}
	}
	if !hasMyRepo {
		t.Fatalf("Children = %v, want to include %q", entry.Children, "my-repo")
	}
	if !hasNotes {
		t.Fatalf("Children = %v, want to include %q", entry.Children, "notes")
	}
}
```
