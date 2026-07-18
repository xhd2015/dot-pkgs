# Scenario

**Bug**: server must initiate a clean WS close when the shell exits under an
active writer/attacher

```
short-lived session (still running at attach)
  -> claim writer/attacher
  -> child exits
  -> server-initiated WebSocket end (observe close code and/or Attach Wait)
```

## Preconditions

- Child command is still running when the first attach claims writer so
  `ServeSessionWebSocket` takes the `<-s.done` branch (not `alreadyExited`).
- Default harness command: `sh -c sleep 1`.

## Steps

1. Grouping documents the server-initiated shell-exit path.
2. Leaves set `Phase` to the observation surface under test.

## Context

MECE split at this level is by **how** we observe the close (raw code vs client
Attach Wait), not by different exit triggers. Client-initiated close-1000 churn
is covered elsewhere (lifecycle leak trees) and is out of scope here.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Shared path: shell exit while attached; leaves choose Phase.
	if len(req.Command) == 0 {
		req.Command = []string{"sh", "-c", "sleep 1"}
	}
	return nil
}
```
