# Scenario

**Feature**: pure mouse package APIs for inline TUI hit-testing and origin measurement

```
# app paints frame, registers Hits; terminal emits absolute mouse + CPR
App Hits + abs click -> Resolve / HitTest -> Hit ID (action)

# CSI 6n rides in paint buffer; CPR peeled from input stream
Tracker.FrameSuffix -> CSI 6n -> Terminal CPR -> OriginFromCPR / Tracker.OnCPR
DemuxCPR: peel CPR events, forward SGR mouse and other bytes
```

## Preconditions

- Package `github.com/xhd2015/dot-pkgs/go-pkgs/tui/mouse` is importable.
- All scenarios are pure (no TTY, no Filter Read, no wrk). Leaves set `req.Op`
  and the fields that Op needs.
- Shared hit fixtures match production-shaped Run chips (Y rows 3–4, 4–5, 11–12;
  X 61–71) used by unit tests.

## Context

- Coverage backfill: library behavior is already correct; doctests seal contracts.
- Unit expansion edges (BottomOriginY, OriginFromCPRClamped, ParseCPR malformed,
  Resolve miss, FrameSuffix once) stay in `mouse_test.go` — see root DOCTEST.md
  unit matrix.

## Steps

1. Branch Setup sets `req.Op` and shared geometry.
2. Leaf Setup fills concrete coordinates / CPR values / tracker steps.
3. Root Run dispatches on `req.Op` to the pure API.
4. Leaf Assert checks response fields for that contract.

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/tui/mouse"
)

// leftRunHits is the two-chip row used by HitTest leaves (left + run).
func leftRunHits() []mouse.Hit {
	return []mouse.Hit{
		{Y0: 3, Y1: 4, X0: 1, X1: 61, ID: "left"},
		{Y0: 3, Y1: 4, X0: 61, X1: 71, ID: "run"},
	}
}

// midPaneHits is the mid-pane Run stage stack (add / gen / tag).
func midPaneHits() []mouse.Hit {
	return []mouse.Hit{
		{Y0: 3, Y1: 4, X0: 61, X1: 71, ID: "add-changes"},
		{Y0: 4, Y1: 5, X0: 61, X1: 71, ID: "gen-commit-msg"},
		{Y0: 11, Y1: 12, X0: 61, X1: 71, ID: "tag-next"},
	}
}

// dualGenTagHits is dual-origin fixture: gen row and tag row only.
func dualGenTagHits() []mouse.Hit {
	return []mouse.Hit{
		{Y0: 4, Y1: 5, X0: 61, X1: 71, ID: "gen-commit-msg"},
		{Y0: 11, Y1: 12, X0: 61, X1: 71, ID: "tag-next"},
	}
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	// Root only supplies helpers; branches set Op and inputs.
	if req == nil {
		return fmt.Errorf("nil request")
	}
	return nil
}
```
