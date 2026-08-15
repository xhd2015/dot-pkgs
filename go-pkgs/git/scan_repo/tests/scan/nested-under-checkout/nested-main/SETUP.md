# Scenario

**Feature**: Scan from a linked worktree root discovers nested main repos inside it

```
# consumer-wt is scan root; vendor/nested has its own .git directory (separate repo)
consumer-wt/.git gitlink -> Scan(consumer-wt) -> consumer-wt + vendor/nested rows
```

## Steps

1. Create consumer main + linked worktree `consumer-wt`.
2. Add nested main repo at `vendor/nested` with a `.git` directory inside the worktree.
3. Set `req.Roots` to `consumer-wt`.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := t.TempDir()
	consumerMain := filepath.Join(root, "consumer-main")
	consumerWt := filepath.Join(root, "consumer-wt")
	mkdirAll(t, consumerWt)
	fakeGitWorktree(t, consumerMain, consumerWt)

	nestedMain := filepath.Join(consumerWt, "vendor", "nested")
	fakeGitRepo(t, nestedMain)

	req.Roots = []string{consumerWt}
	return nil
}
```