# Scenario

**Feature**: `--no-cache` discovers repos but writes no durable cache under `--cache-dir`

```
workspace + --no-cache + --cache-dir C
  -> discovers repos on stdout
  -> C has no home/repos.json, no walk.jsonl, no mirror/
```

## Steps

1. Create workspace with one main repo.
2. Args: `--root`, `--cache-dir`, `--no-cache`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	repoDir := filepath.Join(root, "my-repo")
	mkdirAll(t, repoDir)
	fakeGitRepo(t, repoDir)
	cacheDir := t.TempDir()
	req.Args = []string{
		"--root", root,
		"--cache-dir", cacheDir,
		"--no-cache",
	}
	return nil
}
```
