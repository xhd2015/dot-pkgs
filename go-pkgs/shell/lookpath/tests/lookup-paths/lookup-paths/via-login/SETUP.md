# Scenario

**Feature**: names still missing after cheap stages resolve via login shells

```
file stages all miss
  -> RunLogin(bash|zsh, …) for remaining names
  -> From = "bash" | "zsh" (shell basename only)
```

## Steps

1. Force LookPath miss; empty ExtraDirs/candidates/Home.
2. Set default shell order bash then zsh.
3. Leaves configure LoginStdout / LoginFail / LoginStdoutByName.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.LookPathHits = nil
	req.ExtraDirs = nil
	req.ExtraCandidates = nil
	req.Home = ""
	req.Shells = []string{"bash", "zsh"}
	return nil
}
```
