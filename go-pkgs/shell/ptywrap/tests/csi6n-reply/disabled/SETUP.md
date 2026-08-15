# Scenario

**Feature**: `PTYWRAP_NO_DSR_REPLY=1` disables DSR/CPR auto-reply

```
# kill switch
PTYWRAP_NO_DSR_REPLY=1
  -> maybeAutoReplyDSR no-ops
  -> no write; nextPartial nil
```

## Preconditions

- Leaves set `req.DisableEnv = true` (Run applies env via `t.Setenv`).
- Phase uses maybe (write path) so disable is observable on write count.

## Steps

1. Mark disable path; leaves feed a complete `ESC[6n`.
2. Assert zero writes and nil/empty rest.

## Context

Parallel to `PTYWRAP_NO_OSC_REPLY` for OSC. Independent: OSC replies still
work when only DSR is disabled (not exercised in this pure tree).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.DisableEnv = true
	req.Phase = "maybe"
	return nil
}
```
