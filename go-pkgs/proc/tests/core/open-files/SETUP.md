# Scenario

**Feature**: `OpenFiles(pid, opts)` uses `Options.OpenFiles` inject when set

```
Options.OpenFiles = fixture hook -> OpenFiles(pid, opts) -> paths as-is
```

## Preconditions

- Leaves set `req.OpenFilesPID` and `req.OpenFilesInject`.
- No live `lsof`.

## Steps

1. Set `req.Op` to `"open-files"`.
2. Leaf provides pid + paths.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "open-files"
	return nil
}
```
