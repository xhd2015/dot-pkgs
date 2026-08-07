# Scenario

**Feature**: injected runner failure → error

```
Runner -> error -> VerifyScriptable error
```

## Steps

1. Set `ScriptableFail=true`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ScriptableFail = true
	return nil
}
```
