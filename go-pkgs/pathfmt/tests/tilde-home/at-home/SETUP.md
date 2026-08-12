# Scenario

**Feature**: absolute home directory displays as `"~"`

```
# home rules
abs path == home -> "~"
```

## Steps

1. Set `req.Path` to `os.UserHomeDir()`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Path = mustUserHome(t)
	return nil
}
```
