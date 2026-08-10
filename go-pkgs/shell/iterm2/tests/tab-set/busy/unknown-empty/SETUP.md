# Scenario

**Feature**: empty or failed probe is classified unknown

```
ClassifyBusyFromComm("", true)     -> BusyStateUnknown
ClassifyBusyFromComm("zsh", false) -> BusyStateUnknown
ClassifyBusyFromComm("", false)    -> BusyStateUnknown
```

## Steps

1. Leaf Assert covers empty comm and `ok=false` cases.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {

	req.FgComm = ""
	req.FgOK = false
	return nil
}
```
