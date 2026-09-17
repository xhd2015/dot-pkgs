# Scenario

**Feature**: a positive line opens the file at that line

```
FileArgs("code", file, 42) -> ["code", "--goto", "file:42"]
```

## Steps

1. Set `Kind=file` and `Line=42`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Kind = "file"
	req.Line = 42
	return nil
}
```
