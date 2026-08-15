# Scenario

**Feature**: split after bare ESC still completes on second chunk

```
# split after ESC
chunk1 = ESC
chunk2 = [6n
  -> no write until chunk2; final write ESC[3;7R
```

## Steps

1. Set `req.Chunks` to `{\x1b}`, then `{[6n}`.
2. Parent already set cursor `(3,7)` and phase maybe-chunks.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Chunks = [][]byte{
		[]byte{0x1b},
		[]byte("[6n"),
	}
	return nil
}
```
