# Scenario

**Feature**: `OpenFileConfig` opens a file at a line

```
Resolve -> Run(["code", "--goto", "file:7"]) -> Result
```

## Steps

1. Set `Kind=file` and `Line=7`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Kind = "file"
	req.Line = 7
	return nil
}
```
