# Scenario

**Feature**: all names missing → Dirs empty

```
Names=["ghost"] all stages miss
  -> Items Missing; Dirs() == []
```

## Steps

1. Single unknown name; fail login shells; no hits.

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
