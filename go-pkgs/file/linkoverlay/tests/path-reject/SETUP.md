# Scenario

**Feature**: reject overlay paths that escape the target (`..` or absolute)

```
Layer.Files with unsafe Path -> Apply -> error (path validation, not stub)
```

## Steps

1. Leaves set a single Files layer with an unsafe path.
2. Assert expects a real path-safety error (stub `"not implemented"` is insufficient).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.UseApplyDirs = false
	return nil
}
```
