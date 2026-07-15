# Scenario

**Feature**: `--cache-dir` cold scan writes mirror entries under the given path

```
workspace/my-repo/.git
  -> RunCLI --root workspace --cache-dir C
  -> stdout lists my-repo
  -> LoadCacheEntry(C, my-repo) ok with is_repo main
```

## Steps

1. Create workspace with one fake main repo `my-repo/`.
2. Allocate empty temp `cacheDir`.
3. Set `req.Args` to `["--root", workspace, "--cache-dir", cacheDir]` (cache on, no `--no-cache`).

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
	}
	return nil
}
```
