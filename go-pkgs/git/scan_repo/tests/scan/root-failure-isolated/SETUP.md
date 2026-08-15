# Scenario

**Feature**: one bad root does not prevent scanning valid roots

```
valid repo root + missing root -> Scan -> 1 Repo row + 1 RootError; err nil
```

## Steps

1. Create workspace with one fake git repo.
2. Set `req.Roots` to `[goodRoot, missingPath]`.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := t.TempDir()
	good := filepath.Join(root, "good")
	mkdirAll(t, good)
	fakeGitRepo(t, good)
	missing := filepath.Join(root, "does-not-exist")
	req.Roots = []string{good, missing}
	return nil
}
```