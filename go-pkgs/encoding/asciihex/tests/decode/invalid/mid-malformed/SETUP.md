# Scenario

**Feature**: Decode rejects a mid-string step that is not `\x`

```
# first step \x41; next four chars xx42 are not \xHH
Decode(`\x41xx42`) -> error "malformed hex escape sequence at position 4"
```

## Steps

1. Set `req.Hex` to `\x41xx42`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Hex = `\x41xx42`
	return nil
}
```
