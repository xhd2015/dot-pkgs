# Scenario

**Feature**: bash login fails; zsh login succeeds

```
RunLogin bash -> error
RunLogin zsh  -> "/login/zsh/mytool"
  -> Path=..., Via=login_shell:zsh
```

## Steps

1. Set `LoginFail["bash"]=true`.
2. Set `LoginStdout["zsh"]` to a synthetic path.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	p := filepath.Join(req.WorkDir, "login", "zsh", "mytool")
	req.LoginFail = map[string]bool{"bash": true}
	req.LoginStdout = map[string]string{"zsh": p + "\n"}
	return nil
}
```
