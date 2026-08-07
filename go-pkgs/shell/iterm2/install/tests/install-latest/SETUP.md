# Scenario

**Feature**: `InstallLatest` orchestration with fake HTTP + injected Home

```
resolve -> download -> extract -> install -> Register -> VerifyInstalled
(SkipScriptable; no real network / osascript / lsregister)
```

## Steps

1. Set `Operation=install-latest`.
2. Leaves configure HTTP fixture and Home-based target.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "install-latest"
	req.SkipScriptable = true
	return nil
}
```
