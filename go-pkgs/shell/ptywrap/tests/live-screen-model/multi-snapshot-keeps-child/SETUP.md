# Scenario

**Feature**: multi-snapshot remains non-destructive after sticky+dirty session

```
# sticky + dirty session (long-lived child)
fixture TUI -> sticky + dirty frames; sleep

# N short-lived snapshot attaches (FetchStatus-style)
attach_mode=snapshot ×3 (frame + close each)

# child and session survive
ProcessAlive=true; SessionListed=true; SnapshotCount>=3
```

## Preconditions

- Child stays running (`sleep` tail of fixture).
- Each snapshot socket fully closed before the next open.
- Aligns with `snapshot-attach/multi-snapshot-keeps-child-alive` but after a
  sticky+dirty paint path (live-screen session state).

## Steps

1. `Phase=live-screen-multi-snapshot-keeps-child`.
2. `AttachMode=snapshot`.
3. `RepeatCount=3`.
4. Light dirty iters (no ring pressure required).

## Context

Regression: live-screen changes must not make snapshot attach claim writer or
`stopChild` on disconnect.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "live-screen-multi-snapshot-keeps-child"
	req.AttachMode = "snapshot"
	req.RepeatCount = 3
	req.DirtyIters = 20
	req.StickyMarker = "STICKY_FOOTER"
	req.PromptMarker = "STICKY_PROMPT"
	return nil
}
```
