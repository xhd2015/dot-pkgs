# Scenario

**Feature**: dual contract for clean attach end when the shell exits while
attached — status-first exit marker on the client, plus server close **1000**

```
# remote-agent bash + exit must not hang / unexpected EOF
create short-lived shell session
  -> writer/attacher WS attach (before child exits)
  -> child exits
  -> Session broadcasts "[Terminal exited]"  # client may end here (status-first)
  -> ServeSessionWebSocket on <-s.done sends WS close 1000 then close
  -> Attach Wait returns nil (marker and/or close 1000)
```

## Preconditions

1. Package `github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap` importable.
2. `ptywrap.RegisterAPI` + `ptytest` harness available.
3. `sh` available for short-lived child (`sh -c sleep 1`).
4. Phases in `ptytest.Run`:
   - `shell-exit-ws-close-code`
   - `shell-exit-attach-wait`
   - `shell-exit-marker-without-close`
   - `shell-exit-hard-drop-without-marker` (negative control)

## Steps

1. Root `Setup` starts ephemeral HTTP test server; sets `ServerBase`.
2. Groupings narrow by end-of-session signal path (clean shell exit vs hard drop).
3. Leaves pick observation surface (CloseCode vs AttachErr) and assert.

## Context

**Dual contract** (both sides matter):

1. **Client** (`readTerminalOutput`): on text containing `[Terminal exited]`,
   return **nil** immediately — hang-proof even if close is lost or never sent.
2. **Server** (`ServeSessionWebSocket` on `<-s.done`): send WS close **1000**
   then close — tidy happy path for all clients; bare `conn.Close()` used to
   yield **1006** / `unexpected EOF`.

Do **not** silence 1006 globally: mid-session hard drops **without** the exit
marker must still surface a non-nil Attach error.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	base, cleanup := startTestServer(t)
	t.Cleanup(cleanup)
	req.ServerBase = base
	return nil
}
```
