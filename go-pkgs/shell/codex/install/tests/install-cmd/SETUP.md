# Scenario

**Feature**: `Install` — run InstallCmd via injectable RunShell

```
Install(RunShell) -> RunShell(InstallCmd) once
```

## Steps

1. Set `Operation=install`.
2. Root Run always injects a recording `RunShell`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "install"
	return nil
}
```
