# Scenario

**Feature**: path equal to cwd displays as `"."`

```
# cwd rules (checked first)
path == cwd -> "."
```

## Steps

1. Set `req.Path` to the current working directory (project root).

```go
import (
	"github.com/xhd2015/doctest/session"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Path = req.BaseDir
	return nil
}
```
