# Scenario

**Feature**: `"~/foo/bar"` expands to `filepath.Join(home, "foo", "bar")`

```
# Expand pipeline
"~/..." -> filepath.Join(home, suffix)
```

## Steps

1. Set `req.Path` to `"~/foo/bar"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Path = "~/foo/bar"
	return nil
}```
