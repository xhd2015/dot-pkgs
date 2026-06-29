# Scenario

**Feature**: `--ignore-dir` skips a directory by normalized full path only

```
# full path ignore — repo under ignored tree not discovered
caller --ignore-dir <abs>/scratch -> Walk skips scratch -> empty stdout
```

## Steps

1. Create `scratch/hidden-repo/` with fake `.git`.
2. Set `req.Args` with `--root` and `--ignore-dir` set to the absolute `scratch` path.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	scratch := filepath.Join(root, "scratch")
	hidden := filepath.Join(scratch, "hidden-repo")
	mkdirAll(t, hidden)
	fakeGitRepo(t, hidden)
	req.Args = []string{"--root", root, "--ignore-dir", absPath(t, scratch)}
	return nil
}
```