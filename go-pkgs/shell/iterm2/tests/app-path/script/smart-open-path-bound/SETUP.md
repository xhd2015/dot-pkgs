# Scenario

**Feature**: BuildScriptApp embeds path-bound tell

```
AppPath=…/Applications/iTerm.app
  -> BuildScriptApp(appPath, dir)
  -> path-bound tell + smart-open body
```

## Steps

1. Phase `build-script-app`.
2. Explicit AppPath; assert tell only (session logic covered elsewhere).

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "build-script-app"
	req.AppPath = filepath.Join("/tmp/iterm2-smart-open-home", "Applications", "iTerm.app")
	req.Dir = "/tmp/iterm2-app-path-smart-open-proj"
	return nil
}
```
