# Scenario

**Feature**: Scan from a linked worktree root discovers nested linked worktrees inside it

```
# wrk checkout is the scan root; external/mydep is a dep linked wt nested inside
consumer-wt/.git gitlink -> Scan(consumer-wt) -> consumer-wt + external/mydep rows
```

Reproduces the bug where `walkRoot` returns `SkipDir` at the scan root, so nested
checkouts under `external/` are never visited when the caller scans from inside a
checkout (as `wrk --status` and `wrk --done` need).

## Steps

1. Create consumer main + linked worktree `consumer-wt`.
2. Create dep main + linked worktree `external/mydep` nested under `consumer-wt`.
3. Set `req.Roots` to `consumer-wt` (scan from inside the checkout, not a parent).

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	consumerMain := filepath.Join(root, "consumer-main")
	consumerWt := filepath.Join(root, "consumer-wt")
	mkdirAll(t, consumerWt)
	fakeGitWorktree(t, consumerMain, consumerWt)

	depMain := filepath.Join(root, "dep-main")
	externalWt := filepath.Join(consumerWt, "external", "mydep")
	mkdirAll(t, externalWt)
	fakeGitWorktree(t, depMain, externalWt)

	req.Roots = []string{consumerWt}
	return nil
}
```