# Scenario

**Feature**: LookPath miss returns error including binary name

```
all stages miss -> LookPath -> error with "mytool"
```

## Steps

1. Full miss fixtures (no LookPathHit, empty dirs, login fail).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.LookPathHit = ""
	req.ExtraDirs = nil
	req.ExtraCandidates = nil
	req.Home = ""
	req.Shells = []string{"bash", "zsh"}
	req.LoginFail = map[string]bool{"bash": true, "zsh": true}
	return nil
}
```
