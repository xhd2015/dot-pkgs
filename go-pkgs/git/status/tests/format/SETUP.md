# Scenario

**Feature**: `Format(Counts, FormatBackup)` renders machine backup status strings

```
Counts -> Format(FormatBackup) -> clean | dirty (N modified, ...)
```

## Steps

1. Set `req.Op` to `"format"`.
2. Set `req.Style` to `status.FormatBackup`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/status"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Op = "format"
	req.Style = status.FormatBackup
	return nil
}
```
