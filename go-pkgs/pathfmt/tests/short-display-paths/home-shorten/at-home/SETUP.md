# Scenario

**Feature**: home directory displays as `"~"` when cwd is elsewhere

```
# home shorten (when not under cwd)
path == home -> "~"
```

## Steps

1. Set `req.Path` to `os.UserHomeDir()`.

```go
import (
	"github.com/xhd2015/doctest/session"
	"os"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	req.Path = home
	return nil
}```
