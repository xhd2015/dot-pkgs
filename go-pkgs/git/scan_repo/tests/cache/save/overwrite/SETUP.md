# Scenario

**Feature**: sequential Saves leave last writer as valid JSON (atomic overwrite)

```
# last writer wins
SaveCacheEntry(A) -> SaveCacheEntry(B) -> LoadCacheEntry -> B
```

## Steps

1. Set `CacheOp` to `overwrite`.
2. Set `req.Entry` (first write) and `req.EntryB` (second write) with distinct
   field values so Load can prove B won.

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Setup(t *testing.T, req *Request) error {
	req.CacheOp = "overwrite"
	req.Entry = scan_repo.CacheEntry{
		Version:      1,
		RefreshedAt:  "2026-07-15T10:00:00Z",
		MtimeNs:      100,
		IsRepo:       false,
		RepoType:     "",
		GitDir:       "",
		Children:     []string{"old"},
		ScanComplete: false,
		OptionsHash:  "first",
	}
	req.EntryB = scan_repo.CacheEntry{
		Version:      1,
		RefreshedAt:  "2026-07-15T11:30:00Z",
		MtimeNs:      999,
		IsRepo:       true,
		RepoType:     "worktree",
		GitDir:       "/Users/xhd2015/Projects/org/saved-repo/.git/worktrees/wt",
		Children:     []string{"new-a", "new-b"},
		ScanComplete: true,
		OptionsHash:  "second",
	}
	return nil
}
```
