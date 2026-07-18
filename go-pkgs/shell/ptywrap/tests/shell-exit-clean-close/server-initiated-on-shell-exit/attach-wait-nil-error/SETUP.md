# Scenario

**Feature**: client Attach Wait returns nil after normal shell exit (no
`terminal closed: unexpected EOF`)

```
create sh -c sleep 1
  -> ptywrap/client AttachWithIO(Wait=true, SkipTTYCheck)
  -> shell exits; marker and/or server close 1000
  -> Attach returns nil (AttachErr empty)
```

## Preconditions

- Phase `shell-exit-attach-wait` uses real `ptywrap/client` status-first path.
- Pipes for stdin/stdout (non-TTY OK with SkipTTYCheck).

## Steps

1. `Phase=shell-exit-attach-wait`.
2. Default short-lived command from parent unless overridden.

## Context

End-to-end user-facing contract matching `remote-agent bash` + `exit`: attach
wait must not print Error after clean shell exit. Success may come from the
**exit marker** (status-first) and/or **close 1000** via
`normalizeTerminalReadError` — both are valid; this leaf only requires nil Wait.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "shell-exit-attach-wait"
	return nil
}
```
