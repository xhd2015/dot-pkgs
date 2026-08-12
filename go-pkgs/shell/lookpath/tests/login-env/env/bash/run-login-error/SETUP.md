# Scenario

**Feature**: bash single env surfaces RunLogin failure as error

```
RunLogin("bash", …) -> error
ResolveBashLoginEnv("FOO", opts) -> non-nil error
```

## Steps

1. Set EnvName=FOO (non-empty so name validation is not the failure mode).
2. Set LoginFail=true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.EnvName = "FOO"
	req.LoginFail = true
	return nil
}
```
