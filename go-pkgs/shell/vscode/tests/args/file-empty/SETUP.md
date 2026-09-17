# Scenario

**Feature**: an empty file is rejected instead of opening a blank argv slot

```
FileArgs("code", "   ", 0) -> error
```

## Steps

1. Set `Kind=file` and a whitespace-only file.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Kind = "file"
	req.File = "   "
	return nil
}
```
