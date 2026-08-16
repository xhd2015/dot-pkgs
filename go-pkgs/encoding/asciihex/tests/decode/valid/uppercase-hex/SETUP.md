# Scenario

**Feature**: Decode accepts uppercase hex digits (`ParseInt` base 16)

```
Decode(`\x4A\x21`) -> []byte("J!")
```

## Steps

1. Set `req.Hex` to `\x4A\x21` (`A` is uppercase).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Hex = `\x4A\x21`
	return nil
}
```
