# Scenario

**Feature**: present bin with latest fetch failure → noop (skip update; no install)

```
LookPath hit + FetchLatest error
  -> Ensure -> Action=noop
  ShellCalls empty (no install, no update)
  FetchLatestCalls >= 1
```

## Steps

1. Present bin; set `EnsureLatestFail=true`.

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
	req.EnsureLatestFail = true
	return nil
}
```
