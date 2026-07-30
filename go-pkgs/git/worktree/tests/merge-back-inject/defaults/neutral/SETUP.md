# Scenario

**Feature**: defaults are neutral — not hard-coded to WRK_HOME / wrk-merge-back

```
# empty injects → tmp under os.TempDir (or equivalent neutral base)
# stash message is not the product string "wrk-merge-back"
empty TmpDir+StashLabel -> MergeBack dirty-diverged -> success
  -> ObservedTmpPath under os.TempDir()
  -> stash history without requiring wrk-merge-back
```

## Preconditions

- Grouping left inject fields empty.
- No process env mutation.

## Steps

1. Run MergeBack with empty inject options.
2. Assert rebased-and-merged.
3. Assert observed tmp path is under `os.TempDir()` (symlink-canon), not under
   `$HOME/.wrk`.
4. Assert stash history does **not** require `"wrk-merge-back"` (must not be the
   hard-coded product default).

## Context

- RED until implementer stops using `resolveWrkWorktreesDir` / `"wrk-merge-back"`
  as unconditional defaults.
- wrk may still pass product values at call sites (out of scope here).

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	t.Helper()
	if req.SourcePath == "" {
		return fmt.Errorf("neutral: ancestor fixture missing SourcePath")
	}
	if req.TmpDir != "" || req.StashLabel != "" {
		return fmt.Errorf("neutral: inject fields must stay empty; TmpDir=%q StashLabel=%q",
			req.TmpDir, req.StashLabel)
	}
	return nil
}
```
