# Scenario

**Feature**: present outdated bin runs update once

```
LookPath hit + local 0.1.0 + latest 0.2.0
  -> Ensure -> Action=update
  ShellCalls == [UpdateCmd] once
  ResultNeedsUpdate true
```

## Steps

1. Present bin; local raw older than injected latest.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.EnsurePresent = true
	req.LookPathMiss = false
	req.EnsureLocalRaw = "codex-cli 0.1.0"
	req.EnsureLatest = "0.2.0"
	req.EnsureLatestFail = false
	return nil
}
```
