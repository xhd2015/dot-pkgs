# Scenario

**Feature**: disabled Style returns Green text unchanged

```
Style{Enabled:false}.Green("hello") -> "hello"
```

## Steps

1. Set Enabled false, Color `"green"`, Text `"hello"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Enabled = false
	req.Color = "green"
	req.Text = "hello"
	return nil
}
```
