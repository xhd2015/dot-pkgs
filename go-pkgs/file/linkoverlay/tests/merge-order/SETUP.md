# Scenario

**Feature**: layer order — later layers win; within a layer Dir then Files

```
ordered layers (Dir and/or Files) -> Apply(target, layers...) -> conflict resolution
```

## Steps

1. Leaves build multi-layer fixtures with intentional path conflicts.
2. Use `Apply` (not ApplyDirs) unless a leaf overrides.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.UseApplyDirs = false
	return nil
}
```
