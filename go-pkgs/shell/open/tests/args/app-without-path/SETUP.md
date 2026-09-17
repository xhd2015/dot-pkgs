# Scenario

**Feature**: an empty path launches the application itself

```
AppArgs("iTerm", "", "darwin") -> ["open", "-a", "iTerm"]
```

## Steps

1. Set `Kind=app`, `GOOS=darwin`, an app name, and no path.

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
	req.AppName = "iTerm"
	req.AppPath = ""
	return nil
}
```
