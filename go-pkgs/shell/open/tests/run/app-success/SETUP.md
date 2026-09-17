# Scenario

**Feature**: a named app opens the path once

```
AppConfig(app, path, cfg) -> Run(["open", "-a", app, path]) -> Result
```

## Steps

1. Set `Kind=app`, `GOOS=darwin`, an app name, and the fixture `Target`.

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
