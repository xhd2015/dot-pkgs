# Scenario

**Feature**: Attach Wait returns a non-nil error when the peer hard-drops
without an exit marker

```
mock: session_id then bare Close (no "[Terminal exited]", no close 1000)
  -> AttachWithIO(Wait=true)
  -> AttachErr non-empty within timeout
```

## Preconditions

- Phase `shell-exit-hard-drop-without-marker` implemented in `ptytest.Run`.

## Steps

1. `Phase=shell-exit-hard-drop-without-marker`.
2. Assert `AttachErr` is non-empty.

## Context

Negative control for the dual contract: only marker / clean close codes
normalize to nil — bare drop still surfaces as attach failure.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "shell-exit-hard-drop-without-marker"
	return nil
}
```
