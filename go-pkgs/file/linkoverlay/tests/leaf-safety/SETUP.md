# Scenario

**Feature**: leaf writes never follow seed symlinks into the base (no write-through)

```
seed leaf symlink into base -> Files replace same path -> base original unchanged
```

## Steps

1. Leaves seed a leaf via Dir, then replace it with Files content.
2. Assert base file bytes are unchanged.

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
