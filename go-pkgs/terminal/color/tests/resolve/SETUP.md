# Scenario

**Feature**: Resolve maps Mode + TTY + noColorEnv to a bool

```
# injectable — no process env
Resolve(Mode, TTY, noColorEnv) -> Enabled
```

## Preconditions

- `req.Op` is `"resolve"`.
- Leaves set `Mode`, `TTY`, and `NoColorEnv`. Always/Never ignore TTY and env; Auto uses them.

## Steps

1. Set `req.Op` to `"resolve"`.
2. `Run` calls `color.Resolve(req.Mode, req.TTY, req.NoColorEnv)`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "resolve"
	return nil
}
```
