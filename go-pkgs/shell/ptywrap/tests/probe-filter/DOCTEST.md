# ptywrap attach probe filter

When the server auto-replies OSC 10/11 and CSI 6n, those **queries must not
reach attach clients**, and the **local terminal’s reports must not enter the
PTY as keystrokes**. Otherwise `remote-agent bash` shows `rgb:0000/…` mid-output
and leftover `11;rgb:0000/0000/00001R` on the next prompt.

# DSN (Domain Specific Notion)

**Participants**

- **Child PTY** — emits termenv-style probes (`ESC]11;?ST` + `ESC[6n`).
- **Server auto-reply** — already writes OSC reports + CPR into the PTY master.
- **Attach client** — `remote-agent bash` / xterm.js; a real terminal answers
  probes it sees on stdout.
- **`stripOutputProbes`** — remove complete queries from bytes broadcast to
  clients and stored in scrollback; hold incomplete prefixes across chunks.
- **`filterInputReports`** — drop OSC 10/11 reports and CPR from client stdin;
  keep arrows, Ctrl-C, and typed keys.

**Behaviors**

- **Output strip (default)** — complete OSC 10/11 `?` queries and CSI 6n are
  omitted from display; surrounding text is kept.
- **Output hold** — incomplete probe prefixes are not forwarded until complete
  (or ruled out).
- **Non-probe OSC** — OSC 0 window title is forwarded unchanged.
- **Input drop (default)** — `ESC]11;rgb:…ST` and `ESC[row;colR` are omitted;
  following keys are delivered.
- **Input preserve** — `ESC[A` (up) and `ESC[1;1H` (CUP) are not CPR; keep them.
- **Kill switch** — when strip/drop flags are false (mirrors
  `PTYWRAP_NO_OSC_REPLY` / `PTYWRAP_NO_DSR_REPLY`), probes and reports pass
  through.

**Product API sealed for tests**

| Helper | Role |
|--------|------|
| `TestExported_StripOutputProbes(partial, data, stripOSC, stripDSR)` | display + rest |
| `TestExported_FilterInputReports(partial, data, dropOSC, dropCPR)` | keep + rest |

## Version

0.0.1

## Decision Tree

```
shell/ptywrap/tests/probe-filter/
├── DOCTEST.md
├── SETUP.md
├── output-strip/                 # queries not forwarded to clients
│   ├── SETUP.md
│   ├── termenv-pair/             # OSC 11 ST + CSI 6n in noise
│   ├── split-osc/                # hold ESC]11; then complete
│   ├── non-probe-osc0/           # window title kept
│   └── kill-switch/              # flags false → pass through
└── input-drop/                   # reports not injected as keys
    ├── SETUP.md
    ├── leftover-pair/            # rgb:0000 + CPR + ls
    ├── keys-preserved/           # arrows + Ctrl-C
    ├── split-report/             # hold then drop, keep trailing keys
    └── kill-switch/              # flags false → pass through
```

Parameter ranking (most → least significant):

1. **Direction** — output strip vs input drop (the two leak surfaces)
2. **Sequence class** — complete probe/report, split, non-match, kill switch

## Test Index

| # | Leaf | Phase | Description |
|---|------|-------|-------------|
| 1 | `output-strip/termenv-pair` | output | notice + OSC11 + 6n + warning → noticewarning |
| 2 | `output-strip/split-osc` | output-chunks | hold then strip; trailing text kept |
| 3 | `output-strip/non-probe-osc0` | output | OSC 0 title unchanged |
| 4 | `output-strip/kill-switch` | output | flags off → original bytes |
| 5 | `input-drop/leftover-pair` | input | rgb:0000 + CPR dropped; `ls\\r` kept |
| 6 | `input-drop/keys-preserved` | input | arrows + Ctrl-C unchanged |
| 7 | `input-drop/split-report` | input-chunks | hold then drop; trailing keys kept |
| 8 | `input-drop/kill-switch` | input | flags off → original bytes |

## How to Run

```sh
cd external/dot-pkgs/go-pkgs   # brought SSOT
doctest vet ./shell/ptywrap/tests/probe-filter
doctest test ./shell/ptywrap/tests/probe-filter/...
```

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap"
)

// Request drives the probe-filter helpers without a full WebSocket session.
type Request struct {
	// Phase: "output", "output-chunks", "input", "input-chunks".
	Phase string

	Data    []byte
	Partial []byte
	Chunks  [][]byte

	StripOSC bool
	StripDSR bool
	DropOSC  bool
	DropCPR  bool
}

// Response is display/keep bytes plus leftover incomplete prefix.
type Response struct {
	Out  []byte
	Rest []byte
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	switch req.Phase {
	case "", "output":
		out, rest := ptywrap.TestExported_StripOutputProbes(req.Partial, req.Data, req.StripOSC, req.StripDSR)
		return &Response{Out: out, Rest: rest}, nil
	case "output-chunks":
		var partial []byte
		var out []byte
		for _, chunk := range req.Chunks {
			part, rest := ptywrap.TestExported_StripOutputProbes(partial, chunk, req.StripOSC, req.StripDSR)
			out = append(out, part...)
			partial = rest
		}
		return &Response{Out: out, Rest: partial}, nil
	case "input":
		keep, rest := ptywrap.TestExported_FilterInputReports(req.Partial, req.Data, req.DropOSC, req.DropCPR)
		return &Response{Out: keep, Rest: rest}, nil
	case "input-chunks":
		var partial []byte
		var keep []byte
		for _, chunk := range req.Chunks {
			part, rest := ptywrap.TestExported_FilterInputReports(partial, chunk, req.DropOSC, req.DropCPR)
			keep = append(keep, part...)
			partial = rest
		}
		return &Response{Out: keep, Rest: partial}, nil
	default:
		return nil, fmt.Errorf("unknown Phase %q", req.Phase)
	}
}
```
