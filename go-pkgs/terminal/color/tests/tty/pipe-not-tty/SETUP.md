# Scenario

**Feature**: os.Pipe writer is not a TTY

```
WriterIsTTY(os.Pipe writer) -> false
```

## Steps

1. Set WriterKind to `"pipe"`. `Run` opens `os.Pipe` and passes the writer.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.WriterKind = "pipe"
	return nil
}
```
