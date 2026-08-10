# Scenario

**Feature**: missing codex bin runs install once; does not fetch latest

```
LookPath miss -> Ensure -> Action=install
  ShellCalls == [InstallCmd] once
  FetchLatestCalls == 0
```

## Steps

1. Set `EnsurePresent=false` (LookPath miss).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.EnsurePresent = false
	req.LookPathMiss = true
	return nil
}
```
