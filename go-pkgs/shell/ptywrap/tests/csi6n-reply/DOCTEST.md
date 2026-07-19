# ptywrap CSI 6n (DSR/CPR) Auto-Reply

Plan phase **P3** contract tests: headless PTYs (tty-watch / doctest) answer
child **CSI Device Status Report cursor** queries `ESC [ 6 n` with
**CPR** `ESC [ <row> ; <col> R` using the live **vt10x** cursor (**1-based**),
in the same architectural slot as OSC 10/11 color auto-reply
(`shell/ptywrap/osc_reply.go`).

Classic TDD: CSI 6n auto-reply is **not** implemented yet. This tree is
**RED** until the implementer adds pure helpers (and session wire-up) that
satisfy the sealed contract below.

# DSN (Domain Specific Notion)

**Participants**

- **Child PTY** — TUI / shell process that emits DSR cursor queries
  (`ESC[6n`) when it needs the current cursor position (common in Bubble Tea
  / termenv-style headless probes).
- **ptywrap session readLoop** — reads master PTY bytes; already auto-replies
  OSC 10/11 via `maybeAutoReplyOSC`; must also auto-reply DSR/CPR via a
  parallel helper (`maybeAutoReplyDSR` / `consumeCSI6nQueries`).
- **Live vt10x screen** — session-owned cell model; cursor is 0-based
  (`cursor.X`, `cursor.Y`); CPR uses **1-based** `row = cursor.Y+1`,
  `col = cursor.X+1` (same conversion as `snapshot.go`).
- **Partial buffer** — incomplete ESC/CSI fragment retained across PTY read
  chunks (mirror OSC `oscPartial`; cap trailing rest ~64 bytes).
- **Kill switch** — `PTYWRAP_NO_DSR_REPLY=1` disables DSR auto-reply only
  (parallel to `PTYWRAP_NO_OSC_REPLY=1`; independent of OSC).

**Behaviors**

- **Detect** — scan child output for complete plain CSI 6n: bytes
  `ESC '[' '6' 'n'` (`\x1b[6n`). Do **not** treat `ESC[?6n` (DEC private) or
  `ESC[5n` (DSR status) as cursor queries.
- **Reply** — for each complete query, append CPR
  `ESC '[' <row> ';' <col> 'R'` with decimal row/col (no leading zeros
  required; standard `strconv` form is fine).
- **Cursor source** — pure helpers accept injected **1-based** `(row, col)`.
  Session wire-up (implementer) must pass cursor **after** applying the
  output chunk to the live VT (or equivalent consistent ordering so CPR
  matches the screen model the child would see).
- **Incomplete** — if a chunk ends mid-sequence (`ESC`, `ESC[`, `ESC[6`),
  keep those bytes as `rest` / next partial; no premature write.
- **Non-match** — other ESC/CSI/text: advance past ESC (or ignore), no CPR,
  no leftover partial for complete non-matching sequences.
- **Disable** — when `PTYWRAP_NO_DSR_REPLY=1`, `maybeAutoReplyDSR` writes
  nothing and returns nil partial (same shape as OSC disable).

**Product API sealed for implementer** (unexported helpers + test hooks):

| Helper | Role |
|--------|------|
| `consumeCSI6nQueries(buf []byte, row, col int) (replies, rest []byte)` | Pure scan; `row`/`col` 1-based CPR coords |
| `maybeAutoReplyDSR(write func([]byte) error, partial, data []byte, row, col int) (nextPartial []byte)` | Chunk merge + write + partial cap; respects kill switch |
| `dsrReplyDisabled() bool` | `PTYWRAP_NO_DSR_REPLY == "1"` |
| `TestExported_ConsumeCSI6nQueries(...)` | Test hook calling consume |
| `TestExported_MaybeAutoReplyDSR(...)` | Test hook calling maybe |

## Version

0.0.2

## Decision Tree

```
shell/ptywrap/tests/csi6n-reply/
├── DOCTEST.md
├── SETUP.md
├── enabled/                              # kill switch off (default)
│   ├── complete-6n/                      # full ESC[6n in one buffer
│   │   ├── origin-cpr/                   # cursor 1;1 → ESC[1;1R
│   │   ├── mid-screen-cpr/               # cursor 5;12 → ESC[5;12R
│   │   ├── multi-query/                  # two 6n → two CPR concatenated
│   │   └── embedded-in-noise/            # text around 6n still replies
│   ├── incomplete-buffer/                # partial across maybe-chunks
│   │   ├── split-after-esc/              # ESC | [6n
│   │   ├── split-after-csi-intro/        # ESC[ | 6n
│   │   └── split-before-final-n/         # ESC[6 | n
│   └── non-match/                        # no CPR reply
│       ├── csi-cup/                      # ESC[H
│       ├── csi-5n-status/                # ESC[5n (not cursor)
│       └── dec-private-6n/               # ESC[?6n
└── disabled/                             # PTYWRAP_NO_DSR_REPLY=1
    └── env-no-dsr-reply/                 # complete 6n → no write
```

Parameter ranking (most → least significant):

1. **Kill switch** — enabled vs `PTYWRAP_NO_DSR_REPLY=1`
2. **Sequence class** — complete 6n / incomplete buffer / non-match
3. **Cursor / framing** — row;col values, multi-query, noise, split points

## Test Index

| # | Leaf | Phase | Description |
|---|------|-------|-------------|
| 1 | `enabled/complete-6n/origin-cpr` | consume | `\x1b[6n` + (1,1) → `\x1b[1;1R`, rest empty |
| 2 | `enabled/complete-6n/mid-screen-cpr` | consume | `\x1b[6n` + (5,12) → `\x1b[5;12R` |
| 3 | `enabled/complete-6n/multi-query` | consume | two queries → two CPR replies concatenated |
| 4 | `enabled/complete-6n/embedded-in-noise` | consume | noise around 6n still yields one CPR |
| 5 | `enabled/incomplete-buffer/split-after-esc` | maybe-chunks | no write until second chunk completes |
| 6 | `enabled/incomplete-buffer/split-after-csi-intro` | maybe-chunks | `ESC[` then `6n` |
| 7 | `enabled/incomplete-buffer/split-before-final-n` | maybe-chunks | `ESC[6` then `n` |
| 8 | `enabled/non-match/csi-cup` | consume | `ESC[H` → no reply, no rest |
| 9 | `enabled/non-match/csi-5n-status` | consume | `ESC[5n` → no CPR |
| 10 | `enabled/non-match/dec-private-6n` | consume | `ESC[?6n` → no CPR |
| 11 | `disabled/env-no-dsr-reply` | maybe | env set → no write, nil rest |

## How to Run

```sh
# from go-pkgs module root
cd external/dot-pkgs-master-2026-07-18-1/go-pkgs   # or the vendored go-pkgs path
doctest vet ./shell/ptywrap/tests/csi6n-reply
doctest test ./shell/ptywrap/tests/csi6n-reply/...
```

Expect **RED** until implementer adds `consumeCSI6nQueries` /
`maybeAutoReplyDSR` (+ `TestExported_*` hooks) and wires `session.readLoop`.

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap"
)

// Request drives pure CSI 6n helpers without a full WebSocket session.
type Request struct {
	// Phase selects the surface under test:
	//   "consume"      — TestExported_ConsumeCSI6nQueries(Data, Row, Col)
	//   "maybe"        — one-shot TestExported_MaybeAutoReplyDSR(Partial+Data)
	//   "maybe-chunks" — sequential MaybeAutoReplyDSR over Chunks
	Phase string

	Data    []byte   // consume / maybe single buffer
	Partial []byte   // maybe: prior incomplete fragment
	Chunks  [][]byte // maybe-chunks: ordered PTY read slices

	// Row, Col are 1-based CPR coordinates (cursor.Y+1, cursor.X+1).
	Row int
	Col int

	// DisableEnv sets PTYWRAP_NO_DSR_REPLY=1 for the duration of Run.
	DisableEnv bool
}

// Response captures replies written / returned and leftover partial.
type Response struct {
	Replies    []byte // consume replies or concatenated write() payloads
	Rest       []byte // leftover incomplete fragment
	WriteCalls int    // how many times write was invoked (maybe paths)
}

func Run(t *testing.T, req *Request) (*Response, error) {
	if req.DisableEnv {
		t.Setenv("PTYWRAP_NO_DSR_REPLY", "1")
	}

	row, col := req.Row, req.Col
	if row <= 0 {
		row = 1
	}
	if col <= 0 {
		col = 1
	}

	switch req.Phase {
	case "", "consume":
		replies, rest := ptywrap.TestExported_ConsumeCSI6nQueries(req.Data, row, col)
		return &Response{Replies: replies, Rest: rest}, nil

	case "maybe":
		var got []byte
		calls := 0
		write := func(b []byte) error {
			calls++
			got = append(got, b...)
			return nil
		}
		rest := ptywrap.TestExported_MaybeAutoReplyDSR(write, req.Partial, req.Data, row, col)
		return &Response{Replies: got, Rest: rest, WriteCalls: calls}, nil

	case "maybe-chunks":
		var got []byte
		calls := 0
		write := func(b []byte) error {
			calls++
			got = append(got, b...)
			return nil
		}
		var partial []byte
		for _, chunk := range req.Chunks {
			partial = ptywrap.TestExported_MaybeAutoReplyDSR(write, partial, chunk, row, col)
		}
		return &Response{Replies: got, Rest: partial, WriteCalls: calls}, nil

	default:
		return nil, fmt.Errorf("unknown Phase %q", req.Phase)
	}
}
```
