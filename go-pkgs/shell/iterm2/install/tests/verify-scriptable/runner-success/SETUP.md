# Scenario

**Feature**: injected runner returns version string

```
Runner -> "3.5.0" -> VerifyScriptable returns "3.5.0"
```

## Steps

1. Set `ScriptableVersion=3.5.0`, `ScriptableFail=false`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ScriptableVersion = "3.5.0"
	req.ScriptableFail = false
	return nil
}
```
