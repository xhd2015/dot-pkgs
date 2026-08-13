# Scenario

**Feature**: enabled Strike wraps with SGR 9 and reset

```
Style{Enabled:true}.Strike("hello") -> "\x1b[9m" + "hello" + "\x1b[0m"
```

## Steps

1. Set Enabled true, Color `"strike"`, Text `"hello"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Enabled = true
	req.Color = "strike"
	req.Text = "hello"
	return nil
}
```
