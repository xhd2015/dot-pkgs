# Scenario

**Feature**: `--max-depth` excludes repos beyond depth limit

```
deep repo at depth 3, --max-depth 2 -> empty stdout
```

## Steps

1. Create `level1/level2/deep-repo/` with fake `.git` at depth 3.
2. Set `req.Args` with `--root` and `--max-depth 2`.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := t.TempDir()
	deep := filepath.Join(root, "level1", "level2", "deep-repo")
	mkdirAll(t, deep)
	fakeGitRepo(t, deep)
	req.Args = []string{"--root", root, "--max-depth", "2"}
	return nil
}
```