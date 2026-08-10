# Scenario

**Feature**: InstallViaUserOpen=false does not invoke Open or ClearQuarantineFn

```
InstallLatest(... InstallViaUserOpen=false, Open=rec, Clear=rec)
  -> success; OpenCalls empty; ClearCalls empty
```

## Steps

1. Leave `InstallViaUserOpen` false (default).
2. Root Run still injects recording Open + ClearQuarantineFn.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.InstallViaUserOpen = false
	req.OpenShouldFail = false
	req.ClearShouldFail = false
	return nil
}
```
