# Scenario

**Feature**: CSI CUP `ESC[H` is not a cursor query — no CPR

```
# non-match CUP
Data = ESC[H
  -> replies empty; rest empty
```

## Steps

1. Set `req.Data` to `\x1b[H`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Data = []byte("\x1b[H")
	return nil
}
```
