# Scenario

**Feature**: ModeFromFlags maps the --color / --no-color pair to Mode

```
# parser-agnostic bool pair
ModeFromFlags(color, noColor) -> Mode or exclusive-or error
```

## Preconditions

- `req.Op` is `"flags"`.
- Leaves set `ColorFlag` and `NoColorFlag`. No `less-flags` in tests or implied API.

## Steps

1. Set `req.Op` to `"flags"`.
2. `Run` calls `color.ModeFromFlags(req.ColorFlag, req.NoColorFlag)`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "flags"
	return nil
}
```
