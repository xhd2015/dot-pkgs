# Scenario

**Feature**: bash login fails; zsh login succeeds; From=zsh

```
RunLogin bash -> error
RunLogin zsh  -> "/login/zsh/mytool"
  -> Item Path=…, From="zsh"
```

## Steps

1. Single name `mytool`.
2. LoginFail bash; LoginStdout zsh path.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Names = []string{"mytool"}
	p := filepath.Join(req.WorkDir, "login", "zsh", "mytool")
	req.LoginFail = map[string]bool{"bash": true}
	req.LoginStdout = map[string]string{"zsh": p + "\n"}
	return nil
}
```
