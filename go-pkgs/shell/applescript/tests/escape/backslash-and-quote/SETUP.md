# Scenario

**Feature**: EscapeString escapes backslash and double-quote only.

```
EscapeString(`echo "hi"\x`) → echo \"hi\"\\x  (see ASSERT)
```

## Steps

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = "escape"
	req.EscapeInput = `echo "hi"\`
	return nil
}
```
