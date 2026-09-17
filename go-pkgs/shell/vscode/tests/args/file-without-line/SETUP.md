# Scenario

**Feature**: a zero line opens the file without `--goto`

```
FileArgs("code", file, 0) -> ["code", file]
```

## Steps

1. Set `Kind=file` and `Line=0`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Kind = "file"
	req.Line = 0
	return nil
}
```
