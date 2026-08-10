# Scenario

**Feature**: good final TERM is not clobbered by spawn policy (S7)

```
# preserves-good
after merge, TERM is present, non-empty, and not dumb
  -> EnsureSpawnTERM
  -> TERM unchanged
```

## Preconditions

- Leaves supply a usable TERM (default or other good value).

## Steps

1. Leaves set Base with a good TERM; empty Set/Unset unless overridden.

## Context

- Includes `xterm-256color` itself and other values such as `screen-256color`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Set = nil
	req.Unset = nil
	return nil
}
```
