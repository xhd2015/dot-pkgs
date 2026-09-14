# Scenario

**Feature**: client OSC 10/11 reports and CPR are dropped from PTY input

```
# input-drop
dropOSC + dropCPR true
  -> rgb reports and ESC[row;colR omitted
  -> real keys kept
```

## Steps

1. Set phase to `input` (leaves may override to `input-chunks`).
2. Keep drop flags on.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "input"
	req.DropOSC = true
	req.DropCPR = true
	return nil
}
```
