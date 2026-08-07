# Scenario

**Feature**: explode intermediate seed symlinks when a later op needs a directory there

```
Dir seed (.config as symlink) + later write under .config/tool/...
  -> unlink .config, mkdir, re-link children -> sibling + new path both visible
```

## Steps

1. Leaves seed a nested tree via Dir, then write a divergent nested path.
2. Assert both sibling and new paths are readable under target.

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
