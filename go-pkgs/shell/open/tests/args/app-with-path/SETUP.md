# Scenario

**Feature**: `open -a` targets a path with a named application

```
AppArgs("Visual Studio Code", path, "darwin") -> ["open", "-a", app, path]
```

## Steps

1. Set `Kind=app`, `GOOS=darwin`, an app name, and `AppPath`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Kind = "app"
	req.GOOS = "darwin"
	req.AppName = "Visual Studio Code"
	req.AppPath = req.Target
	return nil
}
```
