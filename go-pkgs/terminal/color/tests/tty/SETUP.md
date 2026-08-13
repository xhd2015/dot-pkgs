# Scenario

**Feature**: WriterIsTTY reports whether a writer is a TTY

```
WriterIsTTY(writer) -> false for buffer and pipe
# do not require a real TTY
```

## Preconditions

- `req.Op` is `"tty"`.
- Leaves set `WriterKind` to `"buffer"` or `"pipe"`. `Run` constructs the writer.

## Steps

1. Set `req.Op` to `"tty"`.
2. `Run` builds the writer and calls `color.WriterIsTTY(w)`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "tty"
	return nil
}
```
