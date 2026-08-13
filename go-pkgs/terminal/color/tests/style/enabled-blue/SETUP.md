# Scenario

**Feature**: enabled Blue wraps with SGR 34 and reset

```
Style{Enabled:true}.Blue("hello") -> "\x1b[34m" + "hello" + "\x1b[0m"
```

## Steps

1. Set Enabled true, Color `"blue"`, Text `"hello"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Enabled = true
	req.Color = "blue"
	req.Text = "hello"
	return nil
}
```
