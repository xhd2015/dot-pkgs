# Scenario

**Feature**: Encode of `lgu_AB` is the locked lowercase `\xHH` string with no newline

```
Encode([]byte("lgu_AB")) -> \x6c\x67\x75\x5f\x41\x42
```

## Steps

1. Set `req.Data` to `[]byte("lgu_AB")`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Data = []byte("lgu_AB")
	return nil
}
```
