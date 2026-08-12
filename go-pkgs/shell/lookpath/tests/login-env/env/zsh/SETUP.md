# Scenario

**Feature**: zsh single-var login env via ResolveZshLoginEnv

```
ResolveZshLoginEnv(name, opts)
  -> RunLogin("zsh", …) -> lookup name in dump
```

## Steps

1. Set `Shell=zsh`.
2. Leaves set EnvName and inject fixtures.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Shell = "zsh"
	return nil
}
```
