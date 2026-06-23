# Scenario

**Feature**: missing install targets still shorten relative to cwd

```
# evalPath canonicalizes through longest existing prefix when leaf is missing
missing file under cwd subtree -> Short -> cwd-relative display (not absolute)
```

## Preconditions

- Parent directories may exist while the leaf file does not (typical pre-install paths).
- `evalPath` must resolve cwd and target through matching canonical prefixes even when
  `filepath.EvalSymlinks` fails on the missing leaf (e.g. macOS `/var` vs `/private/var`).

## Steps

1. Save and restore the original cwd.
2. Create a temp project directory; optionally create parent dirs without the leaf file.
3. `chdir` to the project root.
4. Leaves set `req.Path` to an absolute path whose leaf file is missing.

## Context

- `req.Op` defaults to `"short"`.
- Expectations use platform-native `filepath` separators.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	saveAndRestoreCwd(t)
	projRoot := t.TempDir()
	req.Path = projRoot
	chdirTo(t, projRoot)
	return nil
}
```