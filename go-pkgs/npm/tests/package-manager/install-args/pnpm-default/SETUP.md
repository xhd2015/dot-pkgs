# Scenario

**Feature**: pnpm default install argv for InstallArgs

```
# install argv builder
Manager pnpm + FrozenLockfile false -> install
```

## Steps

1. Set `req.Manager` and `req.FrozenLockfile`.

```go
import (
	"testing"

	npm "github.com/xhd2015/dot-pkgs/go-pkgs/npm"
)

func Setup(t *testing.T, req *Request) error {
	req.Manager = npm.ManagerPnpm
	req.FrozenLockfile = false
	return nil
}
```