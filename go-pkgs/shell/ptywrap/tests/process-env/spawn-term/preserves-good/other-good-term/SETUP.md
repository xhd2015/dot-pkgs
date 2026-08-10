# Scenario

**Feature**: non-default good TERM (screen-256color) is preserved (S7)

```
# other good TERM
Base has TERM=screen-256color and PATH=/bin
  -> EnsureSpawnTERM
  -> TERM still screen-256color (not forced to xterm-256color)
```

## Steps

1. Base: `TERM=screen-256color`, `PATH=/bin`.
2. Empty Set and Unset.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Base = []string{
		"TERM=screen-256color",
		"PATH=/bin",
	}
	req.Set = nil
	req.Unset = nil
	return nil
}
```
