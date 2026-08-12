# Scenario

**Feature**: bash empty GOPATH → zsh login hit

```
RunLogin("bash", …) <- (no GOPATH / empty)
RunLogin("zsh", …)  <- GOPATH=/tmp/from-zsh\0
ResolveGoPathWith -> ("/tmp/from-zsh", nil)
```

## Steps

1. Bash dump without GOPATH (successful empty).
2. Zsh dump with `GOPATH=/tmp/from-zsh` (set by parent group).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	// Successful bash dump with no GOPATH → ("", nil) from ResolveBashLoginEnv.
	req.BashStdout = nulEnvDump("PATH=/usr/bin")
	req.BashFail = false
	return nil
}
```
