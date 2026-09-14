# Scenario

**Feature**: arrow keys and Ctrl-C are not mistaken for CPR

```
# keys
ESC[A ESC[B ESC[C ESC[D Ctrl-C
  -> unchanged
```

## Steps

1. Set `req.Data` to arrows plus Ctrl-C.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Data = []byte("\x1b[A\x1b[B\x1b[C\x1b[D\x03")
	return nil
}
```
