# Scenario

**Feature**: final TERM missing, empty, or dumb → default `xterm-256color` (S6)

```
# needs-default
after merge, TERM is missing | "" | dumb
  -> EnsureSpawnTERM
  -> TERM=xterm-256color
```

## Preconditions

- Leaves arrange for a bad/missing TERM after merge.

## Steps

1. Leaves set Base (and optionally Unset/Set) so final pre-policy TERM is unusable.

## Context

- Empty means `TERM=` (key present, empty value).
- `dumb` means exactly the string `dumb`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	// Shared non-TERM marker so policy does not wipe unrelated keys.
	if req.Base == nil {
		req.Base = []string{"PATH=/bin"}
	}
	return nil
}
```
