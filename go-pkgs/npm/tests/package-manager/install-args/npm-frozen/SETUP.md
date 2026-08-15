# Scenario

**Feature**: npm frozen install argv uses ci

```
# install argv builder
Manager npm + FrozenLockfile true -> ci
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
	req.Manager = npm.ManagerNpm
	req.FrozenLockfile = true
	return nil
}
```