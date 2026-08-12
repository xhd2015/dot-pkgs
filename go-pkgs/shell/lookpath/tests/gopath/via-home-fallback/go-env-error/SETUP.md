# Scenario

**Feature**: RunGoEnv error is soft → continue to ~/go

```
bash/zsh empty GOPATH
LookPath("go") -> /fake/go
RunGoEnv("/fake/go") -> error
ResolveGoPathWith -> (filepath.Join(Home, "go"), nil)
```

## Steps

1. LookPath succeeds with GoBin.
2. Set GoEnvFail=true.

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
	req.GoEnvFail = true
	return nil
}
```
