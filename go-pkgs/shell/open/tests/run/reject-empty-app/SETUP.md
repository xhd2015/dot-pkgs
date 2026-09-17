# Scenario

**Feature**: an empty app name is rejected before the runner

```
AppConfig("", path, cfg) -> error
runner not called
```

## Steps

1. Set `Kind=app`, `GOOS=darwin`, a blank app name, and a real path.

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
	req.AppName = "  "
	req.AppPath = req.Target
	return nil
}
```
