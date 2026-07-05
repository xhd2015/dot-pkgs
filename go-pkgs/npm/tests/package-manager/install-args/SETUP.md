# Scenario

**Feature**: `InstallArgs` and `InstallCommand` build per-manager install argv

```
# manager + frozen flag
Manager + FrozenLockfile -> InstallArgs -> argv slice
```

## Steps

1. Leaf `Setup` sets `req.Manager` and `req.FrozenLockfile`.
2. `req.Op` is `install-args`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Op = "install-args"
	return nil
}
```