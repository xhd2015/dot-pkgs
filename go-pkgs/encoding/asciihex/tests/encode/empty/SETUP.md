# Scenario

**Feature**: Encode of empty bytes is the empty string

```
Encode([]byte{}) -> ""
```

## Steps

1. Set `req.Data` to an empty non-nil slice.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Data = []byte{}
	return nil
}
```
