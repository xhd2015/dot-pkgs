# Scenario

**Feature**: incomplete CSI 6n fragments buffer across PTY read chunks

```
# maybe-chunks
chunk1 ends mid-sequence -> no write; partial retained
chunk2 completes ESC[6n -> write ESC[r;cR; partial cleared
```

## Preconditions

- `req.Phase = "maybe-chunks"`.
- Leaves supply ordered `req.Chunks`.

## Steps

1. Set phase to maybe-chunks.
2. Leaves define split points of `\x1b[6n`.
3. Assert no premature reply after first chunk; final replies match CPR.

## Context

Mirrors OSC `maybeAutoReplyOSC` partial behavior (`osc_reply.go`).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "maybe-chunks"
	// Known cursor for CPR once sequence completes.
	req.Row = 3
	req.Col = 7
	return nil
}
```
