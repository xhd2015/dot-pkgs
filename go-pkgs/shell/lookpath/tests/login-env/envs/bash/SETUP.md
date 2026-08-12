# Scenario

**Feature**: bash full login env dump via ResolveBashLoginEnvs

```
ResolveBashLoginEnvs(opts)
  -> RunLogin("bash", dump-cmd, minimal env)
  -> []string KEY=value
```

## Steps

1. Set `Shell=bash`.
2. Leaves inject NUL dump or LoginFail.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Shell = "bash"
	return nil
}
```
