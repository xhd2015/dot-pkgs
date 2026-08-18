# Scenario

**Feature**: `-v` / `--verbose` prints a warning when a directory is skipped due to permission denied

```
# unreadable child dir during walk
caller --root + -v -> Walk SkipDir on permission error -> stderr warning
```

## Steps

1. Build workspace with `visible-repo` and unreadable `secret` directory.
2. Run CLI with `--root` and `-v`.

```go
import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if runtime.GOOS == "windows" {
		t.Skip("chmod permission fixture requires unix")
	}
	if os.Geteuid() == 0 {
		t.Skip("root bypasses chmod 000; permission-skip warning is not observable")
	}
	root := t.TempDir()
	visible := filepath.Join(root, "visible-repo")
	mkdirAll(t, visible)
	fakeGitRepo(t, visible)
	addUnreadableDir(t, root, "secret")
	req.Args = []string{"--root", root, "-v", "--no-cache"}
	return nil
}
```