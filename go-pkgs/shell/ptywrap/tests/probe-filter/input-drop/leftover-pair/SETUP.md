# Scenario

**Feature**: the leftover `rgb:0000` + CPR pair is dropped; following keys remain

```
# leftover pair (user transcript)
ESC]11;rgb:0000/0000/0000 ST + ESC[20;1R + ls CR
  -> keep ls CR
```

## Steps

1. Set `req.Data` to the local-terminal reply pair plus `ls\r`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Data = []byte("\x1b]11;rgb:0000/0000/0000\x1b\\\x1b[20;1Rls\r")
	return nil
}
```
