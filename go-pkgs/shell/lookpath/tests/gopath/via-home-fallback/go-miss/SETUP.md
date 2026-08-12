# Scenario

**Feature**: go binary miss (LookPath fail) → soft continue to ~/go

```
bash/zsh empty GOPATH
LookPath("go") -> error
ResolveGoPathWith -> (filepath.Join(Home, "go"), nil)
# RunGoEnv not required after miss
```

## Steps

1. Set LookPathFail=true.
2. Do not require RunGoEnv success.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.LookPathFail = true
	req.GoBin = ""
	return nil
}
```
