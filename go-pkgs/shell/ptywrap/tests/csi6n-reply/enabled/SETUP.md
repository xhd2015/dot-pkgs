# Scenario

**Feature**: DSR auto-reply is enabled (default; kill switch unset)

```
# kill switch off
PTYWRAP_NO_DSR_REPLY unset/not-1
  -> complete ESC[6n produces CPR write/reply
  -> incomplete sequences buffer
  -> non-6n sequences ignored
```

## Preconditions

- `req.DisableEnv` remains false.
- Environment does not force `PTYWRAP_NO_DSR_REPLY=1` for these leaves.

## Steps

1. Grouping sets enabled path defaults (no disable env).
2. Child nodes split by sequence class: complete / incomplete / non-match.

## Context

Default production path for headless tty-watch and doctest PTYs.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.DisableEnv = false
	return nil
}
```
