# Scenario

**Feature**: incomplete OSC 11 report is held until ST, then dropped

```
# split report
chunk1 = ESC]11;rgb:0000
chunk2 = /0000/0000 ST + ab
  -> keep ab
```

## Steps

1. Set phase to `input-chunks`.
2. Split the report before the ST terminator.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "input-chunks"
	req.Chunks = [][]byte{
		[]byte("\x1b]11;rgb:0000"),
		[]byte("/0000/0000\x1b\\ab"),
	}
	return nil
}
```
