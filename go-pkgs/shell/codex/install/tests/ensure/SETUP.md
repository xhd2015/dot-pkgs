# Scenario

**Feature**: `Ensure` orchestration — missing install / outdated update / current noop

```
LookPath + LocalVersion + LatestVersion
  -> missing  -> Install  (Action=install); no latest fetch
  -> outdated -> Update   (Action=update)
  -> current  -> noop     (Action=noop); no shell mutator
  -> latest fail when present -> noop (Action=noop); no install
```

## Steps

1. Set `Operation=ensure`.
2. Leaves set EnsurePresent / local raw / latest / fail flags.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "ensure"
	return nil
}
```
