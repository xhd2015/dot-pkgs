# Scenario

**Feature**: attach_mode=snapshot is read-only and does not kill the PTY child

```
# multi-poll snapshot (FetchStatus-style)
create long-lived child
  -> N× WS attach_mode=snapshot (frame + close)
  -> child still running; snapshots non-empty
```

## Preconditions

1. Package `github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap` importable.
2. `ptywrap.RegisterAPI` + `ptytest` harness available.
3. `sleep` / `sh` available for long-lived child.
4. Phase `snapshot-multi-keeps-child` implemented in `ptytest.Run`.

## Steps

1. Root `Setup` starts ephemeral HTTP test server; sets `ServerBase`.
2. Leaf sets `Phase=snapshot-multi-keeps-child`, `AttachMode=snapshot`, `RepeatCount≥3`.
3. Assert child alive and snapshot count / marker.

## Context

Writer path (`attach_mode=screen`) claims writer and may `stopChild` on disconnect —
that breaks multi-poll waitForPrompt + inject (CSI Down / Enter / /status).
Production `ttywatch.ReadSnapshot` must use `attach_mode=snapshot`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	base, cleanup := startTestServer(t)
	t.Cleanup(cleanup)
	req.ServerBase = base
	return nil
}
```
