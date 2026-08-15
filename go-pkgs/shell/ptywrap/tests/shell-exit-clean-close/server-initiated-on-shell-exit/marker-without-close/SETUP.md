# Scenario

**Feature**: Attach Wait ends on the session exit marker even when the peer
never sends a WebSocket close frame (hang-proof status path)

```
mock WS: session_id + "[Terminal exited]", no close
  -> AttachWithIO(Wait=true)
  -> returns nil within 3s (AttachErr empty)
```

## Preconditions

- Phase `shell-exit-marker-without-close` starts its own mock server.
- Root ServerBase is unused by this phase.

## Steps

1. `Phase=shell-exit-marker-without-close`.
2. Assert empty AttachErr.

## Context

Client half of the dual contract: `readTerminalOutput` must not `continue`
past the exit marker and wait forever for close. Proves hang-proofing when
close is lost or never sent. Negative control (hard drop without marker still
errors) lives under `hard-drop-without-marker/`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "shell-exit-marker-without-close"
	return nil
}
```
