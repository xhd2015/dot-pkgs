# Scenario

**Feature**: BuildPathScanSmokeScriptApp embeds path-bound tell

```
AppPath=…/Applications/iTerm.app
  -> BuildPathScanSmokeScriptApp(appPath)
  -> path-bound tell + path scan smoke body
```

Representative “other builder uses same header” leaf (smoke; focus/session-list
may share the same helper in product).

## Steps

1. Phase `build-smoke-app`.
2. Explicit AppPath.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "build-smoke-app"
	req.AppPath = filepath.Join("/tmp/iterm2-smoke-home", "Applications", "iTerm.app")
	return nil
}
```
