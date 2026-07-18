# Scenario

**Feature**: shell exits under an active writer/attacher — observe clean end via
server close 1000 and/or client Attach Wait success (including marker-only)

```
short-lived session (still running at attach)
  -> claim writer/attacher
  -> child exits
  -> exit marker broadcast + server-initiated WS close 1000
  -> leaves observe CloseCode and/or Attach Wait
```

## Preconditions

- Child command is still running when the first attach claims writer so
  `ServeSessionWebSocket` takes the `<-s.done` branch (not `alreadyExited`).
- Default harness command: `sh -c sleep 1`.
- `marker-without-close` uses a mock peer (no real child); root ServerBase is
  unused for that leaf.

## Steps

1. Grouping documents the clean shell-exit attach-end path.
2. Leaves set `Phase` to the observation surface under test.

## Context

MECE split under this grouping is by **observation / signal reliability**:

| Leaf | Signal under test |
|------|-------------------|
| `ws-close-code-1000` | Server sends close **1000** on shell exit |
| `attach-wait-nil-error` | Real end-to-end Attach Wait → nil |
| `marker-without-close` | Client ends on marker alone (no close frame) |

Hard-drop **without** marker (must still error) is a sibling group, not this
path. Client-initiated close-1000 churn is covered elsewhere (lifecycle trees).

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
