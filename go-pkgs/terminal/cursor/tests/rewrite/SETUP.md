# Scenario

**Feature**: Rewrite prefixes CR + ClearLine

```
Rewrite("hi") -> "\r\x1b[2Khi"
```

## Steps

1. Set Op `"rewrite"`, Text `"hi"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "rewrite"
	req.Text = "hi"
	return nil
}
```
