# ptywrap Shell Exit — Clean WebSocket Close (1000)

When a remote shell **exits while a writer/attacher is connected**, the server
must end the WebSocket with **close code 1000 (Normal Closure)** so attach-wait
clients return success. A bare `conn.Close()` without a close frame yields
**1006** / `unexpected EOF` and surfaces as
`Error: terminal closed: unexpected EOF` (e.g. `remote-agent bash` then `exit`).

# DSN (Domain Specific Notion)

**Participants**

- **Session (`ptywrap`)** — owns the PTY child; closes `s.done` when the child
  exits; session may remain listable with `status=exited`.
- **ServeSessionWebSocket** — writer/attacher path: on `<-s.done` must send WS
  close **1000** then close (not bare TCP close).
- **Attach client (`ptywrap/client`)** — `Attach`/`AttachWithIO` with `Wait`
  bridges IO until disconnect; `normalizeTerminalReadError` treats close **1000**
  (and GoingAway / 4000) as **nil**, but leaves **1006** / unexpected EOF as an
  error (`terminal closed: unexpected EOF`).
- **Test harness (`ptytest`)** — starts httptest + `RegisterAPI`, creates a
  short-lived shell session, attaches before exit, records `CloseCode` and
  `AttachErr`.

**Behaviors**

- Shell exits while writer attached → observed WebSocket close code is **1000**.
- Client Attach Wait after shell exit → **nil** error (empty `AttachErr`).
- Session may still appear in `GET /sessions` with `status=exited` (lifecycle
  contract unchanged). Mid-session hard drops (true 1006) are out of scope.

## Version

0.0.2

## Decision Tree

```
go-pkgs/shell/ptywrap/tests/shell-exit-clean-close/
├── DOCTEST.md
├── SETUP.md
└── server-initiated-on-shell-exit/     # shell ends; server closes the attach WS
    ├── SETUP.md
    ├── ws-close-code-1000/             # raw WS close code == 1000
    └── attach-wait-nil-error/          # client Attach Wait returns nil
```

Parameter ranking (most → least significant):

1. **Close initiator / shell-exit path** — server closes after `<-s.done` while
   attached (this tree); not client-initiated disconnect churn.
2. **Observation surface** — raw WebSocket close code vs client Attach Wait
   error normalization.
3. **Shell lifetime** — short-lived child (`sh -c sleep 1`) still running at
   attach time so the done branch runs (not already-exited reattach).

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `server-initiated-on-shell-exit/ws-close-code-1000` | Writer attach; shell exits; CloseCode==1000 |
| 2 | `server-initiated-on-shell-exit/attach-wait-nil-error` | Attach Wait after shell exit returns nil |

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
