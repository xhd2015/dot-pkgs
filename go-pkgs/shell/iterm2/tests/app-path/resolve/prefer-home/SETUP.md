# Scenario

**Feature**: prefer ~/Applications/iTerm.app over /Applications when both exist

```
env unset; home + system both IsApp
  -> ResolveAppPathWith = filepath.Join(home, "Applications", "iTerm.app")
```

## Steps

1. Env unset.
2. Both home and system in ExistingDirs.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.HomeDir = "/tmp/iterm2-resolve-home-prefer"
	req.EnvSet = false
	req.ExistingDirs = []string{
		homeApp(req.HomeDir),
		systemApp,
	}
	return nil
}
```
