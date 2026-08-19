# Scenario

**Feature**: multiple `--root` values union scan results

```
--root A --root B -> repos from both roots, path-sorted lines
```

## Steps

1. Create two workspace roots, each with one fake repo.
2. Set `req.Args` with two `--root` flags.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	parent := t.TempDir()
	rootA := filepath.Join(parent, "scan-a")
	rootB := filepath.Join(parent, "scan-b")
	mkdirAll(t, rootA)
	mkdirAll(t, rootB)
	fakeGitRepo(t, filepath.Join(rootA, "repo-a"))
	fakeGitRepo(t, filepath.Join(rootB, "repo-b"))
	req.Args = []string{"--root", rootA, "--root", rootB, "--no-cache"}
	return nil
}
```