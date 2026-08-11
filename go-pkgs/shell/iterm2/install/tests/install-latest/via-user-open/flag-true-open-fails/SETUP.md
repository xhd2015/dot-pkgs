# Scenario

**Feature**: InstallViaUserOpen=true + Open failure aborts without place

```
InstallLatest(... InstallViaUserOpen=true, Open=error)
  -> error
  -> Open called once on staged path
  -> no ClearQuarantineFn
  -> no Register / no Applications place
```

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
