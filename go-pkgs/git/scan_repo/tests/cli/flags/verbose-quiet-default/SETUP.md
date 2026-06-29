# Scenario

**Feature**: without `-v`, permission-denied directory skips are silent on stderr

```
# same unreadable child as verbose case, no -v flag
caller --root -> Walk SkipDir -> stderr empty
```

## Steps

1. Build workspace with `visible-repo` and unreadable `secret` directory.
2. Run CLI with `--root` only (no verbose flag).

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
	req.Args = []string{"--root", root}
	return nil
}
```