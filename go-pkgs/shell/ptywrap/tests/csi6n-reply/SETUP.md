# Scenario

**Feature**: ptywrap auto-answers CSI 6n (DSR cursor) with CPR from vt10x cursor

```
# headless PTY DSR/CPR auto-reply (parallel to OSC 10/11)
child emits ESC[6n on PTY master read
  -> consumeCSI6nQueries / maybeAutoReplyDSR
  -> write ESC[<row>;<col>R into PTY master (child stdin)
  -> row/col = live vt10x cursor (1-based)

# partial buffer across chunks
chunk1 incomplete ESC/CSI fragment -> rest retained
chunk2 completes 6n -> CPR write; rest cleared

# kill switch
PTYWRAP_NO_DSR_REPLY=1 -> no write, nil partial
```

## Preconditions

1. Package `github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap` is importable.
2. Implementer provides pure helpers + test hooks (RED until present):
   - `TestExported_ConsumeCSI6nQueries(buf []byte, row, col int) (replies, rest []byte)`
   - `TestExported_MaybeAutoReplyDSR(write func([]byte) error, partial, data []byte, row, col int) (nextPartial []byte)`
3. No WebSocket / full session required — tests inject 1-based cursor coordinates.
4. Existing OSC reply unit tests (`osc_reply_test.go`) remain green and unchanged.

## Steps

1. Leaves set `req.Phase`, input bytes (`Data` / `Chunks`), and `Row`/`Col`.
2. Root `Run` calls the sealed test hooks and records `Replies` / `Rest`.
3. Assert compares exact CPR bytes `ESC[r;cR` or emptiness for non-match / disable.

## Context

- **1-based cursor**: product session uses `cursor.Y+1`, `cursor.X+1` (see
  `snapshot.go`); pure helpers take already-converted row/col so tests do not
  depend on vt10x.
- **Session wire-up** (out of pure-test scope, implementer checklist): call
  DSR auto-reply with cursor **after** applying output to the live screen
  model; keep a `dsrPartial` (or shared partial) field next to `oscPartial`.
- **Independence**: `PTYWRAP_NO_DSR_REPLY` does not affect OSC;
  `PTYWRAP_NO_OSC_REPLY` does not affect DSR.
- Mirror OSC rest cap (~64 trailing bytes) for pathological incomplete input.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Defaults: origin cursor, consume phase. Leaves override as needed.
	if req.Row == 0 {
		req.Row = 1
	}
	if req.Col == 0 {
		req.Col = 1
	}
	if req.Phase == "" {
		req.Phase = "consume"
	}
	return nil
}
```
