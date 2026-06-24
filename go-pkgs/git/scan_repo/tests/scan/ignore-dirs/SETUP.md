# Scenario

**Feature**: default ignore skips repos under `node_modules`

```
node_modules/ is ignored -> hidden-repo/.git not discovered
```

## Steps

1. Create `node_modules/hidden-repo/` with `.git`.
2. Set `req.Roots` to workspace (no custom IgnoreDirs).

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	hidden := filepath.Join(root, "node_modules", "hidden-repo")
	mkdirAll(t, hidden)
	fakeGitRepo(t, hidden)
	req.Roots = []string{root}
	return nil
}
```