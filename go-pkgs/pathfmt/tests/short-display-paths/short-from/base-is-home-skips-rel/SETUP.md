# Scenario

**Feature**: when baseDir is home, under-home paths use `~/...` not `.spl/...`

```
baseDir = $HOME
path    = $HOME/.spl/seatalk-local-bot/sessions/sid/SYSTEM.md
ShortFrom -> ~/.spl/...  (NOT .spl/...)
```

## Steps

1. Set BaseDir to user home.
2. Set Path under `~/.spl/...`.

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
	req.BaseDir = home
	req.Path = filepath.Join(home, ".spl", "seatalk-local-bot", "sessions", "sid", "SYSTEM.md")
	return nil
}
```
