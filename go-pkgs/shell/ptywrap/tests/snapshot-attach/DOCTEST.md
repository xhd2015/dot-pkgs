# ptywrap Snapshot Attach — Non-Destructive Multi-Snapshot

Contract tests for `attach_mode=snapshot`: short-lived read-only attaches must
**not** kill the PTY child when the snapshot socket closes. This is the
foundation for multi-poll `FetchStatus` / waitForPrompt loops that snapshot
repeatedly while injecting CSI keys.

# DSN (Domain Specific Notion)

**Participants**

- **Session manager (`ptywrap`)** — owns long-lived PTY child and scrollback.
- **Snapshot attach** — `attach_mode=snapshot` claims `roleSnapshot`: sends one
  frame then closes; does **not** claim writer; writer disconnect must not run
  for this role.
- **Writer / screen attach** — `attach_mode=screen` (or default) claims writer;
  writer normal close may call `stopChild` (reap PTY) — **must not** be used for
  multi-poll snapshot paths.
- **Test harness (`ptytest`)** — starts httptest with `RegisterAPI`, creates a
  long-lived `printf …; sleep 3600` session, performs N snapshot attaches,
  reports `ProcessAlive` + `SnapshotCount`.

**Behaviors**

- After ≥3 `attach_mode=snapshot` connects (each fully closed), child PID still
  alive (`ProcessAlive=true`).
- At least one snapshot returns non-empty usable bytes containing the marker
  (`SnapshotCount >= N` preferred; ≥1 required with marker in `WSOutput`).
- Documented contract: multi-poll snapshot must use `snapshot`, never `screen`.

## Version

0.0.2

## Decision Tree

```
go-pkgs/shell/ptywrap/tests/snapshot-attach/
├── DOCTEST.md
├── SETUP.md
└── multi-snapshot-keeps-child-alive/   # N× snapshot attach; child alive
```

Parameter ranking (most → least significant):

1. **Attach mode** — snapshot (non-destructive) [this tree]
2. **Repeat count** — N≥3 short-lived attaches
3. **Child liveness** — ProcessAlive after disconnects

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `multi-snapshot-keeps-child-alive` | ≥3 snapshot attaches; child PID still alive; snapshots non-empty |

## How to Run

```sh
# from go-pkgs module root (dot-pkgs/go-pkgs)
cd go-pkgs
doctest vet ./shell/ptywrap/tests/snapshot-attach
doctest test ./shell/ptywrap/tests/snapshot-attach/...
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
