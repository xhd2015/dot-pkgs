# ptywrap Shell Exit — Clean Attach End

When a remote shell **exits while a writer/attacher is connected**, attach-wait
clients must return success without hanging. End of session is driven by a
**dual contract**:

1. **Client:** `[Terminal exited]` text marker → Attach Wait returns **nil**
   (status-first; does not require a close frame).
2. **Server:** on `<-s.done`, send WS close **1000** then close (secondary;
   keeps the happy path tidy for all clients).

A bare `conn.Close()` without a close frame used to yield **1006** /
`unexpected EOF` (`Error: terminal closed: unexpected EOF` on
`remote-agent bash` then `exit`). Marker-first end fixes hangs even when close
is lost. Hard drops **without** the exit marker must still error (do not
silence 1006 globally).

# DSN (Domain Specific Notion)

**Participants**

- **Session (`ptywrap`)** — owns the PTY child; closes `s.done` when the child
  exits; broadcasts `[Terminal exited]`; session may remain listable with
  `status=exited`.
- **ServeSessionWebSocket** — writer/attacher path: on `<-s.done` sends WS
  close **1000** then close.
- **Attach client (`ptywrap/client`)** — `Attach`/`AttachWithIO` with `Wait`
  bridges IO; returns **nil** when the exit marker is seen (or on close 1000 /
  GoingAway / 4000). Mid-session hard drops without a marker still error.
- **Test harness (`ptytest`)** — real server short-lived shell + mock servers
  (marker-without-close; hard-drop-without-marker); records `CloseCode` and
  `AttachErr`.

**Behaviors**

- Shell exits while writer attached → observed WebSocket close code is **1000**.
- Client Attach Wait after shell exit → **nil** error (empty `AttachErr`).
- Exit marker alone (no close frame) → Attach Wait still **nil** within a short
  timeout (no hang).
- Hard drop without marker → Attach Wait **non-nil** error.
- Session may still appear in `GET /sessions` with `status=exited` (lifecycle
  contract unchanged).

## Version

0.0.2

## Decision Tree

```
go-pkgs/shell/ptywrap/tests/shell-exit-clean-close/
├── DOCTEST.md
├── SETUP.md
├── server-initiated-on-shell-exit/     # clean shell end; dual contract
│   ├── SETUP.md
│   ├── ws-close-code-1000/             # raw WS close code == 1000
│   ├── attach-wait-nil-error/          # client Attach Wait returns nil
│   └── marker-without-close/           # marker alone ends Attach Wait (no hang)
└── hard-drop-without-marker/           # abrupt drop; no exit semantics
    └── attach-wait-non-nil/            # Attach Wait still errors
```

Parameter ranking (most → least significant):

1. **Session-end signal class** — clean shell exit (marker + server 1000) vs
   mid-session hard drop without marker.
2. **Observation surface** — raw WebSocket close code vs client Attach Wait.
3. **Close reliability** — marker-without-close proves hang-proofing.

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `server-initiated-on-shell-exit/ws-close-code-1000` | Writer attach; shell exits; CloseCode==1000 |
| 2 | `server-initiated-on-shell-exit/attach-wait-nil-error` | Attach Wait after shell exit returns nil |
| 3 | `server-initiated-on-shell-exit/marker-without-close` | Exit marker alone → Attach Wait nil (no close) |
| 4 | `hard-drop-without-marker/attach-wait-non-nil` | Hard drop, no marker → Attach Wait non-nil |

## How to Run

```sh
# from go-pkgs module root
doctest vet ./shell/ptywrap/tests/shell-exit-clean-close
doctest test ./shell/ptywrap/tests/shell-exit-clean-close/...
```

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap/ptytest"
)

type Request = ptytest.Request
type Response = ptytest.Response

func Run(t *testing.T, req *Request) (*Response, error) {
	return ptytest.Run(t, req)
}

func startTestServer(t *testing.T) (base string, cleanup func()) {
	return ptytest.StartTestServer(t)
}
```
