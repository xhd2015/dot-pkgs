# Scenario

**Feature**: BuildForceNewWindowScriptApp embeds path-bound tell

```
AppPath=…/Applications/iTerm.app
  -> BuildForceNewWindowScriptApp(appPath, dir)
  -> tell application (POSIX file "…" as text) …
```

## Steps

1. Phase `build-force-new-app`.
2. Explicit AppPath (resolved-home shape); no FS coupling.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "build-force-new-app"
	req.AppPath = filepath.Join("/tmp/iterm2-force-new-home", "Applications", "iTerm.app")
	req.Dir = "/tmp/iterm2-app-path-force-new-proj"
	return nil
}
```
