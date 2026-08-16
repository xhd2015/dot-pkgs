# Scenario

**Feature**: `ShortEnvFrom` falls back to `TildeHome` when no env alias applies

```
# no-replace branch
no eligible prefix match -> TildeHome(path)
empty/nil env -> TildeHome only (no os.Environ magic)
```

## Preconditions

- Leaves use empty/`nil`/unrelated env or non-matching prefixes.
- Result must never be cwd-relative (`Short` form).

## Steps

1. Set `req.Op = "from"`.
2. Leaves set path and env that do not produce a `$VAR` replacement.

## Context

- Home cases use `os.UserHomeDir()` read-only; outside-home uses temp dirs.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "from"
	return nil
}
```
