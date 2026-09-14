# Scenario

**Feature**: ptywrap strips auto-replied probes from attach output and drops duplicate terminal reports from client input

```
# output
child OSC 11 ? + CSI 6n
  -> stripOutputProbes
  -> not forwarded to attach clients

# input
local terminal OSC 11 rgb: + CPR
  -> filterInputReports
  -> not written to PTY as keys
```

## Preconditions

1. Package `github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap` is importable.
2. Test hooks:
   - `TestExported_StripOutputProbes`
   - `TestExported_FilterInputReports`
3. No WebSocket / full session required — pure helpers with injected flags.
4. Existing OSC/DSR auto-reply tests remain green.

## Steps

1. Leaves set `req.Phase`, bytes (`Data` / `Chunks`), and strip/drop flags.
2. Root `Run` calls the sealed hooks.
3. Assert compares `resp.Out` / `resp.Rest`.

## Context

- Default flags are **on** (production: auto-reply enabled).
- Kill-switch leaves set flags to false (pass-through).
- Session wire-up (readLoop / inputLoop) is out of this tree; helpers are the
  contract the session calls.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Phase == "" {
		req.Phase = "output"
	}
	req.StripOSC = true
	req.StripDSR = true
	req.DropOSC = true
	req.DropCPR = true
	return nil
}
```
