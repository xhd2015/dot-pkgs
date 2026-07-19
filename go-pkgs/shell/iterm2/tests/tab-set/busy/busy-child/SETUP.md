# Scenario

**Feature**: non-shell foreground process is classified busy

```
ClassifyBusyFromComm("node"|"spl"|..., true) -> BusyStateBusy
```

## Steps

1. Probe succeeded; foreground is an application (not a login shell).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.FgOK = true
	req.FgComm = "node"
	return nil
}
```
