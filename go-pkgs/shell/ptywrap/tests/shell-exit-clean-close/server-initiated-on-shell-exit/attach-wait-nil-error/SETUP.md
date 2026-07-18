# Scenario

**Bug**: client Attach Wait must return nil after normal shell exit (no
`terminal closed: unexpected EOF`)

```
create sh -c sleep 1
  -> ptywrap/client AttachWithIO(Wait=true, SkipTTYCheck)
  -> shell exits; server closes WS
  -> Attach returns nil (AttachErr empty)
```

## Preconditions

- Phase `shell-exit-attach-wait` uses real `ptywrap/client` normalize path.
- Pipes for stdin/stdout (non-TTY OK with SkipTTYCheck).

## Steps

1. `Phase=shell-exit-attach-wait`.
2. Default short-lived command from parent unless overridden.

## Context

End-to-end user-facing contract matching `remote-agent bash` + `exit`: attach
wait must not print Error after clean shell exit. Relies on server sending
close 1000 so `normalizeTerminalReadError` returns nil.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "shell-exit-attach-wait"
	return nil
}
```
