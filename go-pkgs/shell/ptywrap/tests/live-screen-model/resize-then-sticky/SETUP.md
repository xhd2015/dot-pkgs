# Scenario

**Feature**: after resize, sticky chrome still present in live-screen snapshot

```
# writer resize to known geometry
WS attach (writer) -> resize cols=100 rows=30 -> session + PTY size updated

# fixture waited for winsize, paints sticky on new bottom rows
fixture TUI -> STICKY_FOOTER at row 30; dirty updates on top

# snapshot uses post-resize dimensions / live cells
attach_mode=snapshot -> WSOutput contains STICKY_FOOTER
```

## Preconditions

- Fixture `--wait-cols/--wait-rows` blocks paint until TIOCGWINSZ matches.
- Harness sends resize before fixture paints sticky.
- Live VT must resize (or equivalent) with session geometry so export matches.

## Steps

1. `Phase=live-screen-resize-then-sticky`.
2. `AttachMode=snapshot`.
3. `ResizeCols=100`, `ResizeRows=30`.
4. Light dirty iters after sticky paint.

## Context

Resize is part of the live-screen invariant: geometry changes apply to both PTY
and the persistent cell model before subsequent exports.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "live-screen-resize-then-sticky"
	req.AttachMode = "snapshot"
	req.ResizeCols = 100
	req.ResizeRows = 30
	req.DirtyIters = 20
	req.StickyMarker = "STICKY_FOOTER"
	req.PromptMarker = "STICKY_PROMPT"
	req.ExpectDirty = true
	return nil
}
```
