# Scenario

**Feature**: injected RunVersion failure surfaces as error

```
RunVersion -> error -> LocalVersion error
```

## Steps

1. Set `VersionCmdFail=true`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.VersionCmdFail = true
	req.LookPathMiss = false
	return nil
}
```
