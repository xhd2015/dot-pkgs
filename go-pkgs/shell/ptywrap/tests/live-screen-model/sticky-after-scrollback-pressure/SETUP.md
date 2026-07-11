# Scenario

**Bug**: sticky chrome is lost when scrollback ring drops early paints (cold replay)

```
# sticky paint once
fixture TUI -> STICKY_FOOTER on bottom row

# dirty-only bytes exceed maxScrollback (256 KiB)
fixture TUI -> DIRTY frames pad until ≥ pressure bytes
  -> scrollback ring no longer contains sticky CSI

# cold replay of ring would miss sticky; live VT must retain cells
attach_mode=snapshot -> WSOutput still contains STICKY_FOOTER
```

## Preconditions

- Production `maxScrollback = 256 KiB` (or pressure set above the active cap).
- Fixture emits ≥ ~320 KiB of dirty-only frames after sticky paint.
- Sticky lines are never rewritten after the initial paint.
- This leaf is the **RED-defining** case against current cold-replay behavior.

## Steps

1. `Phase=live-screen-sticky-after-scrollback-pressure`.
2. `AttachMode=snapshot`.
3. Default `PressureBytes` in Run ≈ 320 KiB (256 KiB + 64 KiB margin).
4. Wait until fixture DONE (ring has advanced), then snapshot.

## Context

Models Grok-like dirty-region TUIs: chrome painted early, later frames only
update the active region. Live iTerm keeps chrome in its cell buffer; truncated
scrollback replay does not.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "live-screen-sticky-after-scrollback-pressure"
	req.AttachMode = "snapshot"
	req.StickyMarker = "STICKY_FOOTER"
	req.PromptMarker = "STICKY_PROMPT"
	// Explicit pressure above production maxScrollback (256 KiB).
	req.PressureBytes = 256*1024 + 64*1024
	req.ExpectDirty = true
	return nil
}
```
