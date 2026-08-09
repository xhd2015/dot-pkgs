# Scenario

**Feature**: empty TERM value → TERM=xterm-256color (S6 empty)

```
# empty TERM
Base has TERM= (empty value) and PATH=/bin
  -> EnsureSpawnTERM
  -> TERM=xterm-256color
```

## Steps

1. Base: `TERM=` (empty), `PATH=/bin`.
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
		"TERM=",
		"PATH=/bin",
	}
	req.Set = nil
	req.Unset = nil
	return nil
}
```
