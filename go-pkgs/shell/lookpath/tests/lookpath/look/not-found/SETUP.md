# Scenario

**Feature**: all resolution stages miss → error including binary name

```
LookPath miss + empty dirs/candidates + RunLogin miss -> error("… mytool …")
```

## Steps

1. Leave all hit fixtures empty; shells return not-found.
2. Leaf asserts error text contains the binary name.

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
	req.Shells = []string{"bash", "zsh"}
	req.LoginStdout = nil
	req.LoginFail = map[string]bool{"bash": true, "zsh": true}
	return nil
}
```
