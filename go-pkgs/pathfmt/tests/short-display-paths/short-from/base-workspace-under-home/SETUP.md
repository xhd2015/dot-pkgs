# Scenario

**Feature**: path under `~/.spl` with base = agent workspace under home → `~/...`

```
baseDir = $HOME/seatalk-local-bot (workspace)
path    = $HOME/.spl/seatalk-local-bot/sessions/sid/SYSTEM.md
ShortFrom -> ~/.spl/seatalk-local-bot/sessions/sid/SYSTEM.md
```

## Steps

1. Set BaseDir to a workspace path under home (need not exist).
2. Set Path to SYSTEM.md under `~/.spl/...`.

```go
import (
	"github.com/xhd2015/doctest/session"
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	req.BaseDir = filepath.Join(home, "seatalk-local-bot")
	req.Path = filepath.Join(home, ".spl", "seatalk-local-bot", "sessions", "sid", "SYSTEM.md")
	return nil
}
```
