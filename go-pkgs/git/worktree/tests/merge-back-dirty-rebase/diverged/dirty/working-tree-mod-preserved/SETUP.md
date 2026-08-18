# Scenario

**Feature**: diverged + dirty tracked-file modification + no-rm → working-tree change preserved after sync

```
# User modified shared.txt at line 1. Rebased commit also touches line 1.
# After reset --mixed HEAD, index has rebased content, working tree has user content.
dirty feat -> tmp rebase -> merge -> reset --mixed -> user modification preserved
```

## Steps

1. The diverged+branch topology is already built by parent `diverged/SETUP.md`.
2. Modify `README.md` in working tree (the same file that main's diverging commit touched).
3. Set `WRK_HOME` to temp dir.
4. Call MergeBack.

```go
import (
	"os"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Make a tracked-file modification that conflicts with what the rebase
	// would produce. README.md was created at "init" and modified by main's
	// diverging commit to "# updated\n". We write different content.
	if err := os.WriteFile(filepath.Join(req.SourcePath, "README.md"), []byte("# user modified\n"), 0644); err != nil {
		return err
	}

	wrkHome := filepath.Join(req.WorkRoot, ".wrk")
	if err := os.MkdirAll(wrkHome, 0755); err != nil {
		return err
	}
	return nil
}
```
