# Scenario

**Feature**: nested linked worktree under `external/` is inside the scanning worktree

```
# worktree/external/mydep is a linked wt with its own git toplevel
# replace => <abs>/external/mydep is still under the scanning worktree root
# must classify as within-worktree (IsIntraRepo=true), not extra-repo
worktree + external linked wt + abs replace -> CheckLocalReplaces -> 0 blocking issues
```

## Preconditions

- External worktree exists under `{worktree}/external/mydep-main-2026-07-04`.
- Consumer `go.mod` uses an absolute replace path (as `wrk --dep` writes).

## Steps

1. Create worktree and dep repos with linked external worktree.
2. Write absolute replace in worktree `go.mod`.
3. Call `replace.CheckLocalReplaces(worktree)`.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	consumerTop, externalPath := setupWrkExternalConsumer(t, req.RootDir)
	writeConsumerReplace(t, consumerTop, externalPath)
	req.RootDir = consumerTop
	return nil
}

```