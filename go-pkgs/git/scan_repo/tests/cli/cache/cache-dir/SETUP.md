# Scenario

**Feature**: `--cache-dir` cold scan seeds durable index under the given path (no mirror)

```
workspace/my-repo/.git
  -> RunCLI --root workspace --cache-dir C
  -> stdout lists my-repo
  -> home/repos.json under C; no C/mirror
```

## Steps

1. Create workspace with one fake main repo `my-repo/`.
2. Allocate empty temp `cacheDir`.
3. Set `req.Args` to `["--root", workspace, "--cache-dir", cacheDir]`.

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
