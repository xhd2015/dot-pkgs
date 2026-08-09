# Scenario

**Feature**: empty set/unset → result equals base (S1)

```
# S1 identity
Base = HOME=... PATH=... FOO=keep
Set=[], Unset=[]
  -> MergeProcessEnv
  -> same map as Base (no forced TERM)
```

## Steps

1. Set `req.Base` to a multi-key synthetic environ including a non-TERM marker.
2. Keep Set and Unset empty.

## Context

- Pure merge must not inject `TERM=xterm-256color` when ApplySpawnTERM is false.
- Base deliberately has no TERM so accidental TERM injection fails the assert.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Base = []string{
		"HOME=/tmp/home",
		"PATH=/bin:/usr/bin",
		"FOO=keep",
	}
	req.Set = nil
	req.Unset = nil
	return nil
}
```
