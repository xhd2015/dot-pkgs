# Scenario

**Feature**: DSR status query `ESC[5n` is not a cursor query — no CPR

```
# non-match DSR status
Data = ESC[5n
  -> replies empty; rest empty
```

## Steps

1. Set `req.Data` to `\x1b[5n`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Data = []byte("\x1b[5n")
	return nil
}
```
