# Scenario

**Feature**: non-empty appPath → quoted-path tell header

```
AppPath=/tmp/home/Applications/iTerm.app
  -> TellApplicationHeader
  -> tell application "/tmp/home/Applications/iTerm.app"
```

## Steps

1. Set AppPath to a fake home install path (no FS required for pure header).

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.AppPath = filepath.Join("/tmp/iterm2-header-home", "Applications", "iTerm.app")
	return nil
}
```
