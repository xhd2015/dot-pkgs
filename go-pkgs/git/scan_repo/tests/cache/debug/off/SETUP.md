# Scenario

**Feature**: `Debug=false` produces no `scan:` spam on stderr

```
# debug off (default quiet)
workspace/my-repo + CacheRoot + Debug=false
  -> Scan (cold with cache enabled — would log if Debug were true)
  -> stderr has zero lines/markers with substring scan:
```

## Preconditions

- Parent default `Debug=false` is explicit here for clarity.
- Cache enabled (`NoCache=false`) so a Debug-on Scan would emit phase logs;
  proving silence is non-vacuous.

## Steps

1. Create workspace with one fake main repo `my-repo/`.
2. Set `req.Roots`, `Debug=false`, `NoCache=false`.

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
	req.Roots = []string{root}
	req.NoCache = false
	req.Debug = false
	return nil
}
```
