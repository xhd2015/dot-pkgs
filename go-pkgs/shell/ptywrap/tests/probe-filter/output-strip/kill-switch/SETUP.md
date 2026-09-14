# Scenario

**Feature**: with strip flags off, probes are forwarded unchanged

```
# kill switch
StripOSC=false StripDSR=false
  -> OSC 11 query + CSI 6n pass through
```

## Steps

1. Disable both strip flags.
2. Set `req.Data` to the termenv pair.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.StripOSC = false
	req.StripDSR = false
	req.Data = []byte("\x1b]11;?\x07\x1b[6n")
	return nil
}
```
