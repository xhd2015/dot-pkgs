# Scenario

**Feature**: Decode rejects a truncated last `\xHH` group

```
# first step ok; leftover \x2 is shorter than 4
Decode(`\x41\x2`) -> error "malformed hex escape sequence at position 4"
```

## Steps

1. Set `req.Hex` to `\x41\x2` (7 characters).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Hex = `\x41\x2`
	return nil
}
```
