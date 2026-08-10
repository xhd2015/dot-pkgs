# Scenario

**Feature**: present current bin is a noop (no shell mutator)

```
LookPath hit + local == latest
  -> Ensure -> Action=noop
  ShellCalls empty
  FetchLatestCalls >= 1
```

## Steps

1. Present bin; local matches latest.

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
	req.EnsureLocalRaw = "codex-cli 0.147.0"
	req.EnsureLatest = "0.147.0"
	req.EnsureLatestFail = false
	return nil
}
```
