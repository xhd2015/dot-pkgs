# Scenario

**Feature**: `Options.Refresh` force-cold library surface (P5)

```
# cold seed then force refresh
caller roots + CacheRoot + Refresh
  -> Scan (cold full walk when Refresh=true)
  -> Result.Repos includes post-seed brand-new repos
```

## Preconditions

- Nested root: does not inherit parent helpers; provides own absPath / fixtures.
- Explicit temp `CacheRoot` only (never `$HOME/.cache`).
- `Options.Refresh` must exist on the library type (RED until implementer adds it).

## Steps

1. Allocate temp `CacheRoot`.
2. Leaves cold-seed, plant brand-new, set `Refresh=true`.

```go
import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func absPath(t *testing.T, p string) string {
	t.Helper()
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(abs)
}

func mkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
}

func fakeGitRepo(t *testing.T, dir string) {
	t.Helper()
	gitDir := filepath.Join(dir, ".git")
	mkdirAll(t, filepath.Join(gitDir, "objects"))
}

// coldSeedScan populates a complete root mirror for warm eligibility.
func coldSeedScan(t *testing.T, roots []string, cacheRoot string) {
	t.Helper()
	if cacheRoot == "" {
		t.Fatal("coldSeedScan: empty cacheRoot")
	}
	if len(roots) == 0 {
		t.Fatal("coldSeedScan: empty roots")
	}
	_, err := scan_repo.Scan(context.Background(), scan_repo.Options{
		Roots:     roots,
		CacheRoot: cacheRoot,
		NoCache:   false,
	})
	if err != nil {
		t.Fatalf("cold seed Scan: %v", err)
	}
	rootPath := absPath(t, roots[0])
	entry, ok, loadErr := scan_repo.LoadCacheEntry(cacheRoot, rootPath)
	if loadErr != nil {
		t.Fatalf("cold seed LoadCacheEntry(root): %v", loadErr)
	}
	if !ok {
		t.Fatalf("cold seed: expected mirror entry for scan root %s", rootPath)
	}
	if !entry.ScanComplete {
		t.Fatalf("cold seed: root ScanComplete=false, want true for warm eligibility")
	}
}

func Setup(t *testing.T, req *Request) error {
	req.CacheRoot = t.TempDir()
	req.NoCache = false
	req.Refresh = false
	return nil
}
```
