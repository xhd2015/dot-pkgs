# Scenario

**Feature**: Decode of Encode output returns the original mixed bytes

```
# includes locked ASCII plus NUL and 0xff (raw byte, not UTF-8 U+00FF)
Encode(data) -> Hex -> Decode -> data
```

## Steps

1. Set `req.Data` to `lgu_AB` plus `0x00` and `0xff`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Data = append([]byte("lgu_AB"), 0x00, 0xff)
	return nil
}
```
