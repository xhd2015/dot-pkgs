# Scenario

**Feature**: zsh single env rejects empty name

```
ResolveZshLoginEnv("", opts) -> error
```

## Steps

1. Set EnvName to empty string.
2. Optional dump unused; empty name is invalid regardless.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.EnvName = ""
	req.LoginStdout = nulEnvDump("FOO=1")
	return nil
}
```
