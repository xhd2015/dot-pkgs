# Scenario

**Feature**: PTY output probes are stripped from attach display

```
# output-strip
stripOSC + stripDSR true
  -> OSC 10/11 queries and CSI 6n omitted
  -> surrounding text kept
```

## Steps

1. Set phase to `output` (leaves may override to `output-chunks`).
2. Keep strip flags on.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "output"
	req.StripOSC = true
	req.StripDSR = true
	return nil
}
```
