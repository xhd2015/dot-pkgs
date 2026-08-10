# Scenario

**Feature**: InstallViaUserOpen=true clears quarantine and opens final app path

```
InstallLatest(... InstallViaUserOpen=true)
  -> success
  -> ClearQuarantineFn(appPath) once
  -> Open(appPath) once
  -> appPath == Result.AppPath (…/iTerm.app)
```

## Steps

1. Set `InstallViaUserOpen=true`.
2. Open and ClearQuarantineFn succeed (recording injectables in root Run).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.InstallViaUserOpen = true
	req.OpenShouldFail = false
	req.ClearShouldFail = false
	return nil
}
```
