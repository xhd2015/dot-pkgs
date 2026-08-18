# Scenario

**Feature**: missing leaf shortens when stored path and cwd differ by /private prefix

```
# macOS: temp path may be /var/... while Getwd returns /private/var/...
missing leaf -> Short -> cwd-relative (not absolute leak)
```

## Steps

1. Build absolute path to a missing file under project root using the temp dir string from `t.TempDir()`.
2. `chdir` already points at the same directory via a separate `filepath.Abs` (may differ in `/private` prefix).

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	projRoot := req.Path
	absRoot, err := filepath.Abs(projRoot)
	if err != nil {
		t.Fatal(err)
	}
	req.BaseDir = absRoot
	req.Op = "short-from"
	req.Path = filepath.Join(projRoot, ".codex", "hooks.json")
	return nil
}
```