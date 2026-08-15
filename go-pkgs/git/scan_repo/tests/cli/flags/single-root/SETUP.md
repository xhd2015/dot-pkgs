# Scenario

**Feature**: single `--root` discovers one main repo

```
--root workspace -> one repo line: path\tmain
```

## Steps

1. Create workspace with `my-repo/` containing fake `.git`.
2. Set `req.Args` to `["--root", <workspace>]`.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := t.TempDir()
	repoDir := filepath.Join(root, "my-repo")
	mkdirAll(t, repoDir)
	fakeGitRepo(t, repoDir)
	req.Args = []string{"--root", root}
	return nil
}
```