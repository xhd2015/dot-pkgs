# Scenario

**Feature**: paths under the cwd subtree display as `"."` or `"rel"` (no `./` prefix)

```
# cwd rules (checked first)
path == cwd -> "." | strict child of cwd -> rel
```

## Preconditions

- A temp project root exists; process cwd is changed to that root before each leaf runs.

## Steps

1. Save and restore the original cwd.
2. Create a temp project directory with subdirectories as needed per leaf.
3. `chdir` to the project root (or nested cwd for `parent-path`).

## Context

- `req.Path` is set by each leaf to the absolute or relative path under test.
- `req.Op` defaults to `"short"`.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	projRoot := t.TempDir()
	mkdirAll(t, filepath.Join(projRoot, "child"))
	mkdirAll(t, filepath.Join(projRoot, "a", "b", "c"))
	req.Path = projRoot
	req.BaseDir = projRoot
	req.Op = "short-from"
	return nil
}```
