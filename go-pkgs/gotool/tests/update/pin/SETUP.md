# Scenario

**Feature**: update.Pin — shared library pin API (classic TDD)

```
# consumer requires old + replace; DepDir module (+ optional tags/Version)
# -> Pin(opts) -> drop replace, set require; PinResult filled; DryRun skips writes
```

## Preconditions

- Root `Setup` ensured `go` and `git` on PATH.
- Product `update.Pin` is not implemented yet — pin leaves expect **RED** until implementer.

## Steps

1. Grouping sets `req.Operation = "pin"` (leaves fill dirs / Version / DryRun).
2. Leaf builds tagged or untagged fixture DepDir + consumer with require+replace.
3. Root `Run` calls `update.Pin` **without** Chdir (cwd-independent).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Operation = "pin"
	return nil
}
```
