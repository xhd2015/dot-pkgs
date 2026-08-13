# Scenario

**Feature**: enabled Yellow wraps with SGR 33 and reset

```
Style{Enabled:true}.Yellow("hello") -> "\x1b[33m" + "hello" + "\x1b[0m"
```

## Steps

1. Set Enabled true, Color `"yellow"`, Text `"hello"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Enabled = true
	req.Color = "yellow"
	req.Text = "hello"
	return nil
}
```
