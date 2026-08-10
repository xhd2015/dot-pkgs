# Scenario

**Feature**: usable ITERM2_APP_PATH wins over home and system

```
ITERM2_APP_PATH=/fake/CustomITerm.app (exists)
home + system also exist
  -> ResolveAppPathWith = /fake/CustomITerm.app
```

## Steps

1. Env set to custom path that IsApp accepts.
2. Home and system also in ExistingDirs (must not win).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.HomeDir = "/tmp/iterm2-resolve-home-env-wins"
	custom := "/fake/CustomITerm.app"
	req.EnvSet = true
	req.EnvValue = custom
	req.ExistingDirs = []string{
		custom,
		homeApp(req.HomeDir),
		systemApp,
	}
	return nil
}
```
