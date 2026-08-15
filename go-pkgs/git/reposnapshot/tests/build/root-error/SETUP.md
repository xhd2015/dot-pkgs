# Scenario

**Feature**: scan root failure becomes synthetic node and RootErrors entry

```
RootErrors[{Root, Error}] -> Build -> synthetic Node + Snapshot.RootErrors
```

## Steps

1. Build manual `scan_repo.Result` with one root error.
2. Use temp workspace as `BaseDir` for rel paths.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := t.TempDir()
	badRoot := filepath.Join(root, "missing-scan-root")
	req.Mode = "manual"
	req.BaseDir = root
	req.ManualResult = scan_repo.Result{
		RootErrors: []scan_repo.RootError{{
			Root:  badRoot,
			Error: "stat: no such file or directory",
		}},
	}
	return nil
}
```
