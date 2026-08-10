# Scenario

**Feature**: no env, no home, no system → empty resolve

```
env unset; ExistingDirs empty
  -> ResolveAppPathWith = ""
```

## Steps

1. Env unset; HomeDir set but no apps exist.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.HomeDir = "/tmp/iterm2-resolve-home-empty"
	req.EnvSet = false
	req.ExistingDirs = nil
	return nil
}
```
