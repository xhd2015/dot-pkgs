# Scenario

**Feature**: bytes.Buffer is not a TTY

```
WriterIsTTY(bytes.Buffer) -> false
```

## Steps

1. Set WriterKind to `"buffer"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.WriterKind = "buffer"
	return nil
}
```
