# Scenario

**Feature**: InstallViaUserOpen=true propagates Open failure

```
InstallLatest(... InstallViaUserOpen=true, Open=error)
  -> err != nil
  -> ClearQuarantineFn may have been called (clear before open)
  -> Open was called once and returned error
```

## Steps

1. Set `InstallViaUserOpen=true`.
2. Set `OpenShouldFail=true` so injected Open returns an error.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.InstallViaUserOpen = true
	req.OpenShouldFail = true
	req.ClearShouldFail = false
	return nil
}
```
