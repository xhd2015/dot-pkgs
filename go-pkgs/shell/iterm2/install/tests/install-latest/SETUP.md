# Scenario

**Feature**: `InstallLatest` orchestration with fake HTTP + injected Home

```
resolve -> download -> extract -> install
  [optional: clear quarantine -> open when InstallViaUserOpen]
  -> Register -> VerifyInstalled
(SkipScriptable; no real network / osascript / lsregister / open / xattr)
```

## Steps

1. Set `Operation=install-latest`.
2. Leaves configure HTTP fixture, Home-based target, and optional
   `InstallViaUserOpen` / open-fail flags.
3. Grouping `via-user-open/` covers flag false/true/open-fails; existing
   `pipeline-fake-http/` stays the baseline happy path.

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
