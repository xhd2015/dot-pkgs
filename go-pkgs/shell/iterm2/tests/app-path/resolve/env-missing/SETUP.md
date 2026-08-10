# Scenario

**Feature**: ITERM2_APP_PATH set but missing → empty (no fallthrough)

```
ITERM2_APP_PATH=/missing/sandbox/iTerm.app (not IsApp)
home + system both exist
  -> ResolveAppPathWith = ""   // localbot strictness; do not use home/system
```

## Steps

1. Env set to a path not listed in ExistingDirs.
2. Home and system are present — must still not fall through.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.HomeDir = "/tmp/iterm2-resolve-home-env-missing"
	req.EnvSet = true
	req.EnvValue = "/missing/sandbox/iTerm.app"
	// Home and system exist but must not be used when env is set-and-unusable.
	req.ExistingDirs = []string{
		homeApp(req.HomeDir),
		systemApp,
	}
	return nil
}
```
