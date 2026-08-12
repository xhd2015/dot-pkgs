# Scenario

**Feature**: zsh single env miss — FOO unset → empty string, nil error

```
RunLogin("zsh", …) <- GOPATH=/tmp/gp\0  (no FOO)
ResolveZshLoginEnv("FOO", opts) -> ("", nil)
```

## Steps

1. Set EnvName=FOO.
2. Inject dump with other keys only (FOO absent).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.EnvName = "FOO"
	req.LoginStdout = nulEnvDump("GOPATH=/tmp/gp")
	return nil
}
```
