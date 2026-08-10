# Scenario

**Feature**: TERM=dumb → TERM=xterm-256color (S6 dumb)

```
# dumb TERM
Base has TERM=dumb and PATH=/bin
  -> EnsureSpawnTERM
  -> TERM=xterm-256color (dumb replaced)
```

## Steps

1. Base: `TERM=dumb`, `PATH=/bin`.
2. Empty Set and Unset.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Base = []string{
		"TERM=dumb",
		"PATH=/bin",
	}
	req.Set = nil
	req.Unset = nil
	return nil
}
```
