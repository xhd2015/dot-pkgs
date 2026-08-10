# Scenario

**Feature**: `IsExecutable(path)` pure filesystem check

```
IsExecutable(path) -> true only for existing regular file with execute bit
```

## Steps

1. Set `Operation=is-executable`.
2. Leaves create fixtures under WorkDir and set `IsExecPath`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "is-executable"
	return nil
}
```
