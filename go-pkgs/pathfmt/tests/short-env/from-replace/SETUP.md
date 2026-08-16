# Scenario

**Feature**: `ShortEnvFrom` replaces path prefix with longest eligible `$NAME`

```
# replace branch
env aliases eligible -> longest segment-boundary match -> $NAME + remainder
```

## Preconditions

- Leaves inject synthetic absolute dirs via `t.TempDir()` into `req.Env`.
- `req.Op` is `from` (default ShortEnvFrom).

## Steps

1. Set `req.Op = "from"`.
2. Leaves create temp prefixes, set `req.Env` and `req.Path`.

## Context

- Temp dirs are typically outside user home so the display is `$VAR`, not `~`.

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
