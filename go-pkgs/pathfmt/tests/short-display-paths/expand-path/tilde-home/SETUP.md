# Scenario

**Feature**: `"~"` expands to the absolute home directory

```
# Expand pipeline
"~" -> home directory
```

## Steps

1. Set `req.Path` to `"~"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Path = "~"
	return nil
}```
