# Scenario

**Feature**: incomplete OSC 11 query is held until the next chunk completes

```
# split OSC
chunk1 = hi + ESC]11;
chunk2 = ?BEL + more
  -> display himore
```

## Steps

1. Set phase to `output-chunks`.
2. Split the query after `ESC]11;`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "output-chunks"
	req.Chunks = [][]byte{
		[]byte("hi\x1b]11;"),
		[]byte("?\x07more"),
	}
	return nil
}
```
