# Scenario

**Feature**: Save then Load preserves all `CacheEntry` v1 fields

```
# round-trip
SaveCacheEntry(entry) -> LoadCacheEntry -> same fields
  (also creates nested mirror dirs under CacheRoot)
```

## Steps

1. Set `CacheOp` to `save-load`.
2. Populate `req.Entry` with non-default values for every schema field:
   version 1, RFC3339 `refreshed_at`, non-zero `mtime_ns`, `is_repo` true,
   `repo_type` `"main"`, absolute `git_dir`, non-empty `children`,
   `scan_complete` true, non-empty `options_hash`.

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Setup(t *testing.T, req *Request) error {
	req.CacheOp = "save-load"
	req.Entry = scan_repo.CacheEntry{
		Version:      1,
		RefreshedAt:  "2026-07-15T12:00:00Z",
		MtimeNs:      1720000000123456789,
		IsRepo:       true,
		RepoType:     "main",
		GitDir:       "/Users/xhd2015/Projects/org/saved-repo/.git",
		Children:     []string{"cmd", "internal", "pkg"},
		ScanComplete: true,
		OptionsHash:  "opts-abc123",
	}
	return nil
}
```
