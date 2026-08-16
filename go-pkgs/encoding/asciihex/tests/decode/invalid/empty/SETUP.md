# Scenario

**Feature**: Decode of empty string errors (kool requires a `\x` prefix)

```
Decode("") -> error "invalid hex escape sequence"
```

## Steps

1. Set `req.Hex` to `""`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Hex = ""
	return nil
}
```
