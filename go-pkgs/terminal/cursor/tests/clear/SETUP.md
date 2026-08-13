# Scenario

**Feature**: Clear is CR + erase line

```
Clear() -> "\r\x1b[2K"
```

## Steps

1. Set Op `"clear"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "clear"
	return nil
}
```
