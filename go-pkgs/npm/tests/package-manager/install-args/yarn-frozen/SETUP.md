# Scenario

**Feature**: yarn frozen install argv for InstallArgs

```
# install argv builder
Manager yarn + FrozenLockfile true -> install --frozen-lockfile
```

## Steps

1. Set `req.Manager` and `req.FrozenLockfile`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	npm "github.com/xhd2015/dot-pkgs/go-pkgs/npm"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Manager = npm.ManagerYarn
	req.FrozenLockfile = true
	return nil
}
```