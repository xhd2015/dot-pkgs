# Scenario

**Feature**: zsh full login env dump via ResolveZshLoginEnvs

```
ResolveZshLoginEnvs(opts)
  -> RunLogin("zsh", dump-cmd, minimal env)
  -> []string KEY=value
```

## Steps

1. Set `Shell=zsh`.
2. Leaves inject NUL dump or LoginFail.

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
