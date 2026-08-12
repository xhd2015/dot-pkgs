# Scenario

**Feature**: stage 1 wins — bash login GOPATH is non-empty after TrimSpace

```
ResolveGoPathWith(opts)
  -> ResolveBashLoginEnv("GOPATH", LoginEnv)
  -> non-empty first segment -> return
  # zsh / LookPath / RunGoEnv not used
```

## Steps

1. Leaves inject bash dump with GOPATH present.
2. Assert short-circuit: no zsh, no LookPath, no RunGoEnv.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	_ = req
	// Grouping: cascade winner is bash login. Leaves set BashStdout.
	return nil
}
```
