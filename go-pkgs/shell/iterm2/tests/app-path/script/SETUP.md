# Scenario

**Feature**: representative Build*App scripts embed TellApplicationHeader

```
appPath + dir -> Build*App -> script starts with path-bound tell
```

## Steps

1. Leaves set Phase and AppPath (explicit; no host FS).
2. Assert path-bound tell; reject bare `"iTerm2"` target.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Default explicit home-style path for path-bound script leaves.
	if req.AppPath == "" {
		req.AppPath = filepath.Join("/tmp/iterm2-script-home", "Applications", "iTerm.app")
	}
	return nil
}
```
