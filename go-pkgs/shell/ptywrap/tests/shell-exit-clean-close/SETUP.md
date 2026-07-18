# Scenario

**Bug**: shell exit while attached ends the WebSocket with bare `conn.Close()`
(1006 / unexpected EOF) instead of close code 1000

```
# remote-agent bash + exit → Error: terminal closed: unexpected EOF
create short-lived shell session
  -> writer/attacher WS attach (before child exits)
  -> child exits (<-s.done)
  -> server must WS close 1000 (not bare Close / 1006)
  -> Attach Wait normalizes 1000 → nil error
```

## Preconditions

1. Package `github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap` importable.
2. `ptywrap.RegisterAPI` + `ptytest` harness available.
3. `sh` available for short-lived child (`sh -c sleep 1`).
4. Phases `shell-exit-ws-close-code` and `shell-exit-attach-wait` implemented in
   `ptytest.Run`.

## Steps

1. Root `Setup` starts ephemeral HTTP test server; sets `ServerBase`.
2. Grouping narrows to server-initiated close on shell exit.
3. Leaves pick observation surface (CloseCode vs AttachErr) and assert.

## Context

Production bug is in `ServeSessionWebSocket` on `case <-s.done:` — bare
`conn.Close()`. Client already treats close 1000 as success via
`normalizeTerminalReadError`. Fix is server-side only (send close 1000 then
close). Do not silence 1006 globally.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	base, cleanup := startTestServer(t)
	t.Cleanup(cleanup)
	req.ServerBase = base
	return nil
}
```
