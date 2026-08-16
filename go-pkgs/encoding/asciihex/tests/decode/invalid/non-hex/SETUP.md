# Scenario

**Feature**: Decode rejects non-hex digits after `\x`

```
Decode(`\xGG`) -> error "invalid hex value GG: " + strconv.ParseInt error
```

## Steps

1. Set `req.Hex` to `\xGG`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Hex = `\xGG`
	return nil
}
```
