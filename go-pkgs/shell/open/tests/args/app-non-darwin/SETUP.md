# Scenario

**Feature**: `open -a` is refused on a platform that has no such flag

```
AppArgs("iTerm", path, "linux") -> error
```

## Steps

1. Set `Kind=app`, `GOOS=linux`, and an app name.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Kind = "app"
	req.GOOS = "linux"
	req.AppName = "iTerm"
	return nil
}
```
