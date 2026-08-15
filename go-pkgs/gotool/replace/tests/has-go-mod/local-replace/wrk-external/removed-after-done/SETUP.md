# Scenario

**Feature**: replace path still under worktree after external worktree removed

```
# wrk --done removes external/* worktree but leaves replace in go.mod
# resolved path still lies under the scanning worktree root (dir may be gone)
# must not be classified as outside-worktree
worktree + abs replace -> remove external wt -> CheckLocalReplaces -> 0 blocking issues
```

## Preconditions

- Worktree `go.mod` still has absolute replace to `{worktree}/external/mydep-main-2026-07-04`.
- The external worktree has been removed (simulating `wrk --done` cascade).

## Steps

1. Create worktree with linked external worktree and absolute replace.
2. Remove the external worktree via `git worktree remove`.
3. Call `replace.CheckLocalReplaces(worktree)`.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	consumerTop, externalPath := setupWrkExternalConsumer(t, req.RootDir)
	writeConsumerReplace(t, consumerTop, externalPath)
	depMain := filepath.Join(req.RootDir, "dep")
	removeExternalWorktree(t, depMain, externalPath)
	req.RootDir = consumerTop
	return nil
}

```