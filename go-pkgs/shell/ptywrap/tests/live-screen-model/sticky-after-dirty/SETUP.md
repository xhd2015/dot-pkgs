# Scenario

**Feature**: sticky chrome survives dirty-region-only updates

```
# sticky paint once on bottom rows
fixture TUI -> STICKY_PROMPT + STICKY_FOOTER

# many dirty frames rewrite only top rows
fixture TUI -> ?2026h CUP top DIRTY_n ?2026l  (× DirtyIters)

# snapshot exports live cells (not only recent dirty bytes)
attach_mode=snapshot -> WSOutput contains STICKY_FOOTER and STICKY_PROMPT
```

## Preconditions

- Default PTY geometry (80×24).
- No scrollback ring pressure (dirty payload stays well under 256 KiB).
- Fixture never repaints sticky lines after the initial paint.

## Steps

1. `Phase=live-screen-sticky-after-dirty`.
2. `AttachMode=snapshot`.
3. `DirtyIters=30` (light dirty churn).
4. Expect sticky markers in the snapshot frame.

## Context

Documents the base live-screen contract. May already pass under cold replay if
sticky sequences remain inside the ring; the pressure leaf is the RED-defining
case.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "live-screen-sticky-after-dirty"
	req.AttachMode = "snapshot"
	req.DirtyIters = 30
	req.StickyMarker = "STICKY_FOOTER"
	req.PromptMarker = "STICKY_PROMPT"
	req.ExpectDirty = true
	return nil
}
```
