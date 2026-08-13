# Scenario

**Feature**: enabled Gray wraps with SGR 90 and reset

```
Style{Enabled:true}.Gray("hello") -> "\x1b[90m" + "hello" + "\x1b[0m"
```

## Steps

1. Set Enabled true, Color `"gray"`, Text `"hello"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Enabled = true
	req.Color = "gray"
	req.Text = "hello"
	return nil
}
```
