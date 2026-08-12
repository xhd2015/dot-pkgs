# Scenario

**Feature**: bash single-var login env via ResolveBashLoginEnv

```
ResolveBashLoginEnv(name, opts)
  -> RunLogin("bash", …) -> lookup name in dump
```

## Steps

1. Set `Shell=bash`.
2. Leaves set EnvName and inject fixtures.

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
