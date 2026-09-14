# Scenario

**Feature**: with drop flags off, reports are forwarded as input

```
# kill switch
DropOSC=false DropCPR=false
  -> rgb report + CPR pass through
```

## Steps

1. Disable both drop flags.
2. Set `req.Data` to the leftover pair.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.DropOSC = false
	req.DropCPR = false
	req.Data = []byte("\x1b]11;rgb:0000/0000/0000\x1b\\\x1b[20;1R")
	return nil
}
```
