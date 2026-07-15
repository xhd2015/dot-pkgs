# Scenario

**Feature**: `--refresh` forces a cold full walk so brand-new repos appear after a warm seed

```
# cold seed via library Scan into C (complete root cache)
workspace/known-repo --seed--> C
plant workspace/brand-new-repo/.git
  -> RunCLI --root workspace --cache-dir C --refresh
  -> stdout lists brand-new-repo AND known-repo
  # without --refresh, warm would omit brand-new-repo
```

## Steps

1. Create workspace with `known-repo/`; cold-seed Scan into temp `cacheDir`.
2. Plant `brand-new-repo/` after the seed (uncached).
3. Set `req.Args` to `["--root", workspace, "--cache-dir", cacheDir, "--refresh"]`.

```go
import (
	"context"
	"path/filepath"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	known := filepath.Join(root, "known-repo")
	mkdirAll(t, known)
	fakeGitRepo(t, known)

	cacheDir := t.TempDir()
	_, err := scan_repo.Scan(context.Background(), scan_repo.Options{
		Roots:     []string{root},
		CacheRoot: cacheDir,
		NoCache:   false,
	})
	if err != nil {
		t.Fatalf("cold seed Scan: %v", err)
	}
	// Sanity: root must be warm-eligible after seed.
	rootPath := absPath(t, root)
	entry, ok, loadErr := scan_repo.LoadCacheEntry(cacheDir, rootPath)
	if loadErr != nil {
		t.Fatalf("cold seed LoadCacheEntry(root): %v", loadErr)
	}
	if !ok || !entry.ScanComplete {
		t.Fatalf("cold seed: root not warm-eligible (ok=%v scan_complete=%v)", ok, entry.ScanComplete)
	}

	brandNew := filepath.Join(root, "brand-new-repo")
	mkdirAll(t, brandNew)
	fakeGitRepo(t, brandNew)

	req.Args = []string{
		"--root", root,
		"--cache-dir", cacheDir,
		"--refresh",
	}
	return nil
}
```
