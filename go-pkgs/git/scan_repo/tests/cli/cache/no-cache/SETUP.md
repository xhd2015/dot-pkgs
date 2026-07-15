# Scenario

**Feature**: `--no-cache` discovers repos but writes no mirror under `--cache-dir`

```
workspace/my-repo/.git
  -> RunCLI --root workspace --cache-dir C --no-cache
  -> stdout lists my-repo
  -> C has no entry.json (mirror absent or empty of entries)
```

## Steps

1. Create workspace with one fake main repo `my-repo/`.
2. Allocate temp `cacheDir`.
3. Set `req.Args` to `["--root", workspace, "--cache-dir", cacheDir, "--no-cache"]`.

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
