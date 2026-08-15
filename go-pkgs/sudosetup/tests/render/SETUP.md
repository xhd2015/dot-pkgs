# Scenario

**Feature**: RenderSudoersLine formats NOPASSWD sudoers entries

```
Manager.RenderSudoersLine -> "<user> ALL=(root) NOPASSWD: <cmd> [<args>]"
```

## Steps

1. Set `Request.Operation = "render"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "render"
	return nil
}
```