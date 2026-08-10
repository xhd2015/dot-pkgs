# Scenario

**Feature**: only system /Applications/iTerm.app present → system path

```
env unset; home missing; system IsApp
  -> ResolveAppPathWith = /Applications/iTerm.app
```

## Steps

1. Env unset; HomeDir set but home app not in ExistingDirs.
2. Only systemApp listed as existing.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.HomeDir = "/tmp/iterm2-resolve-home-system-only"
	req.EnvSet = false
	req.ExistingDirs = []string{systemApp}
	return nil
}
```
