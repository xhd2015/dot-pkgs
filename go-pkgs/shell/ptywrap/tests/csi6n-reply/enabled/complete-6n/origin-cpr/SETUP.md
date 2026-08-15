# Scenario

**Feature**: complete `ESC[6n` with cursor at origin replies `ESC[1;1R`

```
# origin CPR
Data=ESC[6n, Row=1, Col=1
  -> consumeCSI6nQueries
  -> replies ESC[1;1R
```

## Steps

1. Set `req.Data` to `\x1b[6n`.
2. Set cursor to 1-based origin `(1,1)`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Data = []byte("\x1b[6n")
	req.Row = 1
	req.Col = 1
	return nil
}
```
