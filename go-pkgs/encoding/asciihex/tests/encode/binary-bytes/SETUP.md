# Scenario

**Feature**: Encode of NUL / DEL / 0xff uses two lowercase hex digits per byte

```
Encode([]byte{0x00, 0x7f, 0xff}) -> \x00\x7f\xff
```

## Steps

1. Set `req.Data` to `{0x00, 0x7f, 0xff}`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Data = []byte{0x00, 0x7f, 0xff}
	return nil
}
```
