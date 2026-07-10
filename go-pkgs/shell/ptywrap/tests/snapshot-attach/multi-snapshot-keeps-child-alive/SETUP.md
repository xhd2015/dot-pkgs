# Scenario

**Bug**: repeated short snapshot attaches must not reap the PTY child

```
session with long-lived child
  -> Snapshot attach ×3 (attach_mode=snapshot)
  -> ProcessAlive=true; SnapshotCount>=3; WSOutput contains SNAP-MARKER
```

## Preconditions

- Child command prints `SNAP-MARKER` then `sleep 3600`.
- Each snapshot socket is fully closed before the next open.

## Steps

1. `Phase=snapshot-multi-keeps-child`.
2. `AttachMode=snapshot`.
3. `RepeatCount=3`.

## Context

Characterization of the non-destructive snapshot contract used by agent-pro
`pkgs/ttywatch.ReadSnapshot` / FetchStatus multi-poll loops.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "snapshot-multi-keeps-child"
	req.AttachMode = "snapshot"
	req.RepeatCount = 3
	return nil
}
```
