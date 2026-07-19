# Scenario

**Feature**: login shells are classified idle

```
ClassifyBusyFromComm("zsh"|"bash"|"fish"|"sh", true) -> BusyStateIdle
```

## Steps

1. Probe succeeded (`FgOK=true`); comm is a shell name (leaf Assert tables shells).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.FgOK = true
	req.FgComm = "zsh"
	return nil
}
```
