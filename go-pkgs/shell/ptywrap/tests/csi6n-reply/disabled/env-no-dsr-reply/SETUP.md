# Scenario

**Feature**: with `PTYWRAP_NO_DSR_REPLY=1`, complete 6n produces no write

```
# kill switch
DisableEnv=true, Data=ESC[6n, Row=1, Col=1
  -> maybeAutoReplyDSR no-ops
  -> Replies empty; Rest nil/empty; WriteCalls=0
```

## Steps

1. Parent sets `DisableEnv` and phase `maybe`.
2. Feed a complete `\x1b[6n` that would otherwise reply.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Data = []byte("\x1b[6n")
	req.Row = 1
	req.Col = 1
	req.Partial = nil
	return nil
}
```
