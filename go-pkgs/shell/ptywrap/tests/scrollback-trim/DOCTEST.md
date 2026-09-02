# ptywrap scrollback trim — escape-sequence safe cut

Classic TDD for scrollback ring trim: when `len > maxScrollback`, the kept
suffix must **not** begin mid CSI/OSC. Cutting inside Codex `\x1b[?2026h`
must not leave a printable `026h` orphan at the head (idle snap1 instability).

# DSN (Domain Specific Notion)

**Participants**

- **Scrollback ring** — session `[]byte` capped at `maxScrollback` (256 KiB).
- **`trimScrollback(buf, max)`** — pure trim helper used by `session.readLoop`
  (and exit-marker path) instead of raw `buf[len-max:]`.
- **CSI / OSC** — escape sequences starting at `ESC` (`\x1b`); must not be
  split across the trim boundary.

**Behaviors**

- If `len(buf) <= max` → return `buf` unchanged.
- Provisional `cut = len - max`. If `cut` falls **inside** an escape sequence
  that started before `cut`, advance `cut` to just after that sequence ends
  (kept size may be **under** `max`).
- If `cut` is escape-safe → keep `buf[cut:]` (exact `max` when possible).
- Crime scene: cut on the `0` of `\x1b[?2026h` → kept must **not** start with
  `026h`; tail marker after the sequence must remain.

**Product API sealed for implementer**

| Helper | Role |
|--------|------|
| `trimScrollback(buf []byte, max int) []byte` | Pure escape-safe trim |
| `TestExported_TrimScrollback(...)` | Test hook |

## Version

0.0.2

## Decision Tree

```
shell/ptywrap/tests/scrollback-trim/
├── DOCTEST.md
├── SETUP.md
└── mid-escape/
    ├── SETUP.md
    └── csi-2026h-at-cut/                 # LEAF: cut inside ?2026h
        ├── SETUP.md
        └── ASSERT.md
```

Parameter ranking (most → least significant):

1. **Cut position** — mid-escape vs clean boundary
2. **Sequence class** — CSI DEC private `?2026h` (crime scene)

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `mid-escape/csi-2026h-at-cut` | Cut on `0` of `\x1b[?2026h` → no `026h` prefix; `TAIL_MARKER` kept |

## How to Run

```sh
cd external/dot-pkgs/go-pkgs
doctest vet ./shell/ptywrap/tests/scrollback-trim
doctest test ./shell/ptywrap/tests/scrollback-trim/...
```

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap"
)

// Request drives pure trimScrollback without a WebSocket session.
type Request struct {
	Data []byte
	Max  int // trim cap; 0 → use len(Data) semantics only via helper
}

// Response is the trimmed scrollback suffix.
type Response struct {
	Trimmed []byte
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	if req.Max <= 0 {
		return nil, fmt.Errorf("Max must be positive")
	}
	out := ptywrap.TestExported_TrimScrollback(req.Data, req.Max)
	return &Response{Trimmed: out}, nil
}
```
