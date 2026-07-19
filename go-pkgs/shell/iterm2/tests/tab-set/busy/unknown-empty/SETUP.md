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
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.FgComm = ""
	req.FgOK = false
	return nil
}
```
