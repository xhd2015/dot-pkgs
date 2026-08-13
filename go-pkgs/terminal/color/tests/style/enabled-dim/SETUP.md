# Scenario

**Feature**: enabled Dim wraps with SGR 2 and reset

```
Style{Enabled:true}.Dim("hello") -> "\x1b[2m" + "hello" + "\x1b[0m"
```

## Steps

1. Set Enabled true, Color `"dim"`, Text `"hello"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Enabled = true
	req.Color = "dim"
	req.Text = "hello"
	return nil
}
```
