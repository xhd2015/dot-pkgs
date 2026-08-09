# Scenario

**Feature**: non-tilde absolute paths are returned unchanged by `Expand`

```
# Expand passthrough
no ~ prefix -> unchanged
```

## Steps

1. Set `req.Path` to a platform absolute path without a `~` prefix.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Path = "/abs/path"
	return nil
}```
