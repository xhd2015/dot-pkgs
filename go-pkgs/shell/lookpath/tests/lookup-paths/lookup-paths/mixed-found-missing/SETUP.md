# Scenario

**Feature**: mixed found and missing names; LookupPaths still returns nil error

```
Names=["found", "ghost"]
  LookPathHits only "found"
  shells all fail
  -> found item + Missing item; err=nil
```

## Steps

1. Two names; inject LookPath only for the first.
2. Fail both login shells so second name ends Missing.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Names = []string{"found", "ghost"}
	hit := filepath.Join(req.WorkDir, "bin", "found")
	writeExecutable(t, hit)
	req.LookPathHits = map[string]string{"found": hit}
	req.ExtraDirs = nil
	req.ExtraCandidates = nil
	req.Home = ""
	req.Shells = []string{"bash", "zsh"}
	req.LoginFail = map[string]bool{"bash": true, "zsh": true}
	return nil
}
```
