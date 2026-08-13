# Scenario

**Feature**: Paint(Gray, Strike) emits both codes and one reset

```
Style{Enabled:true}.Paint("hello", Gray, Strike) -> "\x1b[90m\x1b[9mhello\x1b[0m"
```

## Steps

1. Set Enabled true, Color `"paint-gray-strike"`, Text `"hello"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Enabled = true
	req.Color = "paint-gray-strike"
	req.Text = "hello"
	return nil
}
```
