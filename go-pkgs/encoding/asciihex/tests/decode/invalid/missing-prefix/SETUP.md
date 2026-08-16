# Scenario

**Feature**: Decode rejects a hex-like string that does not start with `\x`

```
Decode("41") -> error "invalid hex escape sequence"
```

## Steps

1. Set `req.Hex` to `"41"` (plain hex, no `\x` prefix).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Hex = "41"
	return nil
}
```
