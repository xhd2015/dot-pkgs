# Scenario

**Feature**: warm second Scan does not create or grow `mirror/`

```
cold Scan seeds index (+ historically mirror)
plant unit-elsewhere/brand-new (soft-omit target)
  -> warm Scan(NoCache=false)
  -> Result may soft-omit brand-new
  -> <CacheRoot>/mirror still absent after warm
```

## Steps

1. Cold-seed one known repo under `unit-a/known-repo`.
2. Plant uncached `unit-elsewhere/brand-new-repo`.
3. Run warm Scan (parent CacheRoot; NoCache=false).

```go
import (
	"context"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := t.TempDir()
	known := filepath.Join(root, "unit-a", "known-repo")
	mkdirAll(t, known)
	fakeGitRepo(t, known)

	req.Roots = []string{root}
	req.NoCache = false
	_, err := scan_repo.Scan(context.Background(), scan_repo.Options{
		Roots:     req.Roots,
		CacheRoot: req.CacheRoot,
		NoCache:   false,
	})
	if err != nil {
		t.Fatalf("cold seed: %v", err)
	}

	brandNew := filepath.Join(root, "unit-elsewhere", "brand-new-repo")
	mkdirAll(t, brandNew)
	fakeGitRepo(t, brandNew)
	return nil
}

func fakeGitRepo(t *testing.T, dir string) {
	t.Helper()
	gitDir := filepath.Join(dir, ".git")
	mkdirAll(t, filepath.Join(gitDir, "objects"))
}
```
