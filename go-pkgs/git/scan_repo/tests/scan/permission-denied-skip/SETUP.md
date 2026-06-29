# Scenario

**Feature**: `Scan` continues when `filepath.WalkDir` hits permission denied on a child directory

```
# unreadable sibling does not abort discovery
caller roots -> Scan -> Walk SkipDir on EACCES -> visible-repo row
```

## Steps

1. Create `visible-repo` with fake `.git`.
2. Create unreadable `secret` directory under the same root.
3. Set `req.Roots` to the workspace root.

```go
import (
	"path/filepath"
	"runtime"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if runtime.GOOS == "windows" {
		t.Skip("chmod permission fixture requires unix")
	}
	root := t.TempDir()
	visible := filepath.Join(root, "visible-repo")
	mkdirAll(t, visible)
	fakeGitRepo(t, visible)
	addUnreadableDir(t, root, "secret")
	req.Roots = []string{root}
	return nil
}
```