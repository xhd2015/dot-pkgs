# Scenario

**Feature**: enabled Red wraps with SGR 31 and reset

```
Style{Enabled:true}.Red("hello") -> "\x1b[31m" + "hello" + "\x1b[0m"
```

## Steps

1. Set Enabled true, Color `"red"`, Text `"hello"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Enabled = true
	req.Color = "red"
	req.Text = "hello"
	return nil
}
```
