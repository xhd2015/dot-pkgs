# Scenario

**Feature**: empty inject fields use product-neutral library defaults

```
# no TmpDir, no StashLabel, no WRK_HOME env
MergeBackOptions{} inject empty -> dirty-diverged still works with neutral defaults
```

## Steps

1. Build diverged dirty fixture.
2. Leave `req.TmpDir` and `req.StashLabel` empty.
3. Leaf asserts success and neutrality of defaults.

## Context

- Does **not** set `WRK_HOME` / `WRK_DATE`.
- Does **not** pass wrk product strings.
- Observed tmp path must not require `~/.wrk` or `WRK_HOME`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	setupDivergedDirty(t, req)
	req.TmpDir = ""
	req.StashLabel = ""
	return nil
}
```
