# Scenario

**Feature**: enabled Green wraps with SGR 32 and reset

```
Style{Enabled:true}.Green("hello") -> "\x1b[32m" + "hello" + "\x1b[0m"
```

## Steps

1. Set Enabled true, Color `"green"`, Text `"hello"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Enabled = true
	req.Color = "green"
	req.Text = "hello"
	return nil
}
```
