# Scenario

**Feature**: no found paths → DirsEnv is empty string

```
Names=["ghost"] all miss -> DirsEnv == ""
```

## Steps

1. Single missing name; fail logins; no hits.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Names = []string{"ghost"}
	req.LookPathHits = nil
	req.ExtraDirs = nil
	req.ExtraCandidates = nil
	req.Home = ""
	req.Shells = []string{"bash", "zsh"}
	req.LoginFail = map[string]bool{"bash": true, "zsh": true}
	return nil
}
```
