# Scenario

**Feature**: InstallViaUserOpen=true opens staged unzipped app (no place)

```
InstallLatest(... InstallViaUserOpen=true)
  -> success
  -> Open(extractedAppPath) once
  -> no ClearQuarantineFn
  -> no Register / no InstallApp to Home Applications
  -> AppPath == staged …/iTerm.app under cache extract
```

## Steps

1. Set `InstallViaUserOpen=true`.
2. Open succeeds (recording injectables in root Run).

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
