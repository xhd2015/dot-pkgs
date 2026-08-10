# Scenario

**Feature**: bare name resolved via injectable `RunLogin` → `Via=login_shell:<shell>`

```
file stages all miss
  -> RunLogin(bash|zsh, "command -v mytool", env)
  -> first success stdout path -> Via=login_shell:<shell>
```

## Steps

1. Force LookPath miss; empty ExtraDirs/candidates/Home.
2. Leaves configure LoginStdout / LoginFail maps.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Name = "mytool"
	req.LookPathHit = ""
	req.ExtraDirs = nil
	req.ExtraCandidates = nil
	req.Home = ""
	// Explicit default shell order for login stage.
	req.Shells = []string{"bash", "zsh"}
	return nil
}
```
