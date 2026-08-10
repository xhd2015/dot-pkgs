# Scenario

**Feature**: `Update` — run UpdateCmd via injectable RunShell

```
Update(RunShell) -> RunShell(UpdateCmd) once
```

## Steps

1. Set `Operation=update`.
2. Root Run always injects a recording `RunShell`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "update"
	return nil
}
```
