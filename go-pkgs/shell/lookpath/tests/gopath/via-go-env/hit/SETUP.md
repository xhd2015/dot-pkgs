# Scenario

**Feature**: both login empty → go binary + go env GOPATH hit

```
RunLogin bash/zsh <- no GOPATH
LookPath("go") -> /fake/go
RunGoEnv("/fake/go") <- "/tmp/from-go\n"
ResolveGoPathWith -> ("/tmp/from-go", nil)
```

## Steps

1. Set GoBin to a fake go path.
2. Set GoEnvStdout to a single GOPATH (with trailing newline as real go env may).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.GoBin = "/fake/go"
	req.LookPathFail = false
	// Production RunGoEnv trims; include whitespace to pin TrimSpace behavior.
	req.GoEnvStdout = "  /tmp/from-go\n"
	req.GoEnvFail = false
	return nil
}
```
