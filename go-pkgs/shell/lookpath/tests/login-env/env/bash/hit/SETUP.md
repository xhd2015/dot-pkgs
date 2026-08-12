# Scenario

**Feature**: bash single env hit — FOO present in dump → value "1"

```
RunLogin("bash", …) <- FOO=1\0GOPATH=/tmp/gp\0
ResolveBashLoginEnv("FOO", opts) -> ("1", nil)
```

## Steps

1. Set EnvName=FOO.
2. Inject NUL dump containing FOO=1 and GOPATH=/tmp/gp.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.EnvName = "FOO"
	req.LoginStdout = nulEnvDump("FOO=1", "GOPATH=/tmp/gp")
	return nil
}
```
