# Scenario

**Feature**: TERM=xterm-256color already set → do not clobber (S7)

```
# good default present
Base has TERM=xterm-256color and PATH=/bin
  -> EnsureSpawnTERM
  -> TERM still xterm-256color
```

## Steps

1. Base: `TERM=xterm-256color`, `PATH=/bin`.
2. Empty Set and Unset.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Base = []string{
		"TERM=xterm-256color",
		"PATH=/bin",
	}
	req.Set = nil
	req.Unset = nil
	return nil
}
```
