# Scenario

**Feature**: enabled Bold wraps with SGR 1 and reset

```
Style{Enabled:true}.Bold("hello") -> "\x1b[1m" + "hello" + "\x1b[0m"
```

## Steps

1. Set Enabled true, Color `"bold"`, Text `"hello"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Enabled = true
	req.Color = "bold"
	req.Text = "hello"
	return nil
}
```
