# Scenario

**Feature**: mid-session hard drop **without** `[Terminal exited]` must still
yield a non-nil Attach Wait error (do not silence 1006 globally)

```
mock WS: session_id only, then bare TCP/WS tear-down (no exit marker, no 1000)
  -> AttachWithIO(Wait=true)
  -> AttachErr non-empty (unexpected EOF / terminal closed / similar)
```

## Preconditions

- Phase `shell-exit-hard-drop-without-marker` starts its own mock server.
- Root ServerBase is unused by this phase.

## Steps

1. Grouping sets default `Phase` for hard-drop observation.
2. Leaf asserts non-empty `AttachErr`.

## Context

MECE complement to `server-initiated-on-shell-exit/marker-without-close`:

| Path | Marker | Close | Attach Wait |
|------|--------|-------|-------------|
| marker-without-close | yes | none | **nil** |
| hard-drop-without-marker | no | abnormal | **non-nil** |

Confirms status-first success is keyed off the exit marker, not “any EOF is OK”.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Hard-drop path: mock peer only; Phase is the harness selector.
	if req.Phase == "" {
		req.Phase = "shell-exit-hard-drop-without-marker"
	}
	return nil
}
```
