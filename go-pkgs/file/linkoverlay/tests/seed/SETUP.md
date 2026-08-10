# Scenario

**Feature**: Dir-only seeding via `ApplyDirs`

```
# fixture base dirs under WorkingDir
bases (top-level names) -> ApplyDirs(target, bases...) -> abs symlinks at target top
```

## Steps

1. Leaves set `UseApplyDirs` and layer fixtures for each base directory.
2. `materializeLayers` writes base trees before `Run`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.UseApplyDirs = true
	return nil
}
```
