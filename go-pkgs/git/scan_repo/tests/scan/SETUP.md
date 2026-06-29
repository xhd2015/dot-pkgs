# Scenario

**Feature**: filesystem walk discovers git repos without git subprocesses

```
# discovery only — no enrichment
caller roots -> Scan -> Walk -> repo rows (ListRemotes=false, ListWorktrees=false)
```

## Preconditions

- Fake `.git` fixtures suffice; real `git` is not required.
- `ListRemotes` and `ListWorktrees` remain false for every leaf in this branch.

## Steps

1. Disable enrichment flags.
2. Provide `fakeGitRepo` and `fakeGitWorktree` helpers for fixture layout.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.ListRemotes = false
	req.ListWorktrees = false
	return nil
}

func fakeGitRepo(t *testing.T, dir string) {
	t.Helper()
	gitDir := filepath.Join(dir, ".git")
	mkdirAll(t, filepath.Join(gitDir, "objects"))
}

func fakeGitWorktree(t *testing.T, mainDir, wtDir string) {
	t.Helper()
	fakeGitRepo(t, mainDir)
	wtName := filepath.Base(wtDir)
	wtGitDir := filepath.Join(mainDir, ".git", "worktrees", wtName)
	mkdirAll(t, wtGitDir)
	absWtGitDir := absPath(t, wtGitDir)
	writeFile(t, filepath.Join(wtDir, ".git"), "gitdir: "+absWtGitDir+"\n")
}

func addUnreadableDir(t *testing.T, root, name string) {
	t.Helper()
	secret := filepath.Join(root, name)
	mkdirAll(t, secret)
	if err := os.Chmod(secret, 0000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(secret, 0755)
	})
}
```