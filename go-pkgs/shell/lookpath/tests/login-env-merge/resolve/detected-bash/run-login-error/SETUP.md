# Scenario

**Feature**: detected bash surfaces RunLogin failure as error (no zsh fallback)

```
DetectShell -> bash
RunLogin("bash") -> error
ResolveLoginEnvs -> error (zsh not tried)
```

## Steps

1. Set BashFail=true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.BashFail = true
	return nil
}
```
