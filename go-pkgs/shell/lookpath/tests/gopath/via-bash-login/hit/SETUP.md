# Scenario

**Feature**: bash login hit — single GOPATH path returned; zsh/go not called

```
RunLogin("bash", …) <- GOPATH=/tmp/from-bash\0
ResolveGoPathWith -> ("/tmp/from-bash", nil)
# no zsh, no LookPath("go"), no RunGoEnv
```

## Steps

1. Inject bash env dump with `GOPATH=/tmp/from-bash`.
2. Leave zsh/go fixtures empty (must not be consulted).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.BashStdout = nulEnvDump("GOPATH=/tmp/from-bash")
	return nil
}
```
