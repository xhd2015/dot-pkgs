# Scenario

**Feature**: all cascade stages empty → ~/go fallback with injected Home

```
bash/zsh empty GOPATH
LookPath("go") -> /fake/go
RunGoEnv -> "" (empty after trim)
ResolveGoPathWith -> (filepath.Join(Home, "go"), nil)
```

## Steps

1. LookPath succeeds so RunGoEnv is reached.
2. GoEnvStdout empty → soft continue to home fallback.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.GoBin = "/fake/go"
	req.LookPathFail = false
	req.GoEnvStdout = ""
	req.GoEnvFail = false
	return nil
}
```
