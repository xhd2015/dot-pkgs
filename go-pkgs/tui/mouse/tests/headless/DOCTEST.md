# tui/mouse headless — fixture geometry via **tty-watch binary**

## Version
0.0.2

A small **fixture-inline** process paints a known-geometry UI with `tui/mouse`
Tracker + hitmap. Headless doctests drive sessions by **executing** the
`tty-watch` **binary** (`run --detach`, `send`, `snapshot`, `kill`) — **no**
Go module import of `github.com/xhd2015/tty-watch` (avoids cycle with
`go-pkgs/shell/ptywrap`). Pure package leaves under `../` stay a separate root.

# DSN (Domain Specific Notion)

### Participants

- **Fixture (`cmd/fixture-inline`)** — long-lived child under a tty-watch PTY.
  Flags: `--anchor=top|mid|bottom` (default `mid`). Prints pad lines then a
  fixed UI frame, embeds CSI 6n via `mouse.Tracker.FrameSuffix`, demuxes CPR +
  SGR mouse on stdin, resolves clicks with `mouse.Resolve` / known origin, and
  emits **machine protocol** lines on stdout (captured in session scrollback).
- **Machine protocol (sealed stdout lines)** — implementer must match exactly:
  - `ORIGIN=<n> VIEW=<v>` after a successful CPR / origin lock (`n` = 0-based
    originY, `v` = painted viewLines).
  - `HIT id=<id> localY=<y> kind=<k>` after a resolved click (`k` is
    `known` | `top` | `bottom`). Misses do not print HIT.
- **Hitmap (view-local, sealed)** — half-open rows, chip X band:
  - `btn-a`: localY **3**, X **2–12** (click col **5**)
  - `btn-b`: localY **4**, X **2–12** (click col **5**)
  - View frame is **5** lines (`VIEW=5`): lines 0–2 non-hit chrome; 3–4 chips.
- **Anchor pad** — blanks **before** the 5-line UI so origin differs:
  - `top`: pad 0 → origin near 0 after CPR
  - `mid`: pad ~8 (default mid-pane)
  - `bottom`: pad enough that UI sits near the bottom of a 24-row PTY
    (≈ `height - viewLines` blanks)
- **tty-watch host (binary dependency)** — all host ops via `exec.Command`:
  `run --detach`, `send --click|--query-cursor`, `snapshot`, `kill`.
  Coord flags on CLI are **0-based**. Resolve binary: `TTY_WATCH_BIN` → PATH →
  `go build` from `TTY_WATCH_DIR` / sibling brought tree (not go.mod require).
- **Harness** — per-leaf temp `Home` (`TTY_WATCH_HOME`), session-scoped fixture
  build cache, poll snapshot for markers, inject click / query-cursor, kill on
  teardown.
- **PTY / CSI 6n** — child needs host CPR auto-reply (ptywrap DSR). Leaves
  labeled `needs-pty,slow`.

### Behaviors

1. **Detach run** — start fixture with `--anchor=<a>`; unique session id per leaf.
2. **Paint wait** — poll `snapshot` until UI marker (`btn-a` / `fixture-inline`)
   or `ORIGIN=` appears (timeout → fail, not silent skip).
3. **Query-cursor** — `send --query-cursor --json` → host VT 0-based row/col;
   after mid paint, cursor sits on last UI line-ish
   (`originY + viewLines - 1` ±1) **or** snapshot already has `ORIGIN=`.
4. **Click** — `absY = originY + localY`, `absX = 5`; inject SGR click; poll for
   `HIT id=…`.
5. **btn-b not a** — click localY=4 → HIT id=btn-b; must not be btn-a.

### Fixture product (implementer; not fully built by designer)

Path: `tui/mouse/cmd/fixture-inline` (package main). Acceptable simple loop
(non-Bubble Tea):

1. Parse `--anchor`.
2. Print pad newlines for anchor.
3. Print 5 UI lines + Tracker `FrameSuffix(height, viewLines)`.
4. Raw/read stdin: DemuxCPR / Filter; OnCPR → print `ORIGIN=… VIEW=…`.
5. On SGR press: Resolve with OriginY; print `HIT id=… localY=… kind=…`.
6. Keep alive until SIGTERM/kill.

Designer ships a **minimal stub** that prints pad + chrome and sleeps so the
tree builds and leaves **RED** on missing `ORIGIN=` / `HIT=` until the real
fixture lands.

## Decision Tree

```
headless/                              # nested root (own Run contract)
├── mid/                               # --anchor=mid (default geometry)
│   ├── query-cursor-after-paint/      # ORIGIN line and/or cursor on last UI line
│   ├── click-btn-a/                   # abs click localY=3 → HIT btn-a
│   └── click-btn-b-not-a/             # localY=4 → HIT btn-b (not btn-a)
├── top/                               # --anchor=top
│   └── click-btn-a/                   # origin near 0; still HIT btn-a
└── bottom/                            # --anchor=bottom
    └── click-btn-a/                   # bottom pad; still HIT btn-a
```

Parameter ranking (most → least significant):

1. **Anchor** — top / mid / bottom pad changes originY
2. **Action** — query-cursor vs click target chip

## Test Index

| Leaf | Labels | Description |
|------|--------|-------------|
| `mid/query-cursor-after-paint` | needs-pty, slow | After mid paint: `ORIGIN=` present and/or query-cursor near last UI line |
| `mid/click-btn-a` | needs-pty, slow | Mid: click chip A → `HIT id=btn-a localY=3` |
| `mid/click-btn-b-not-a` | needs-pty, slow | Mid: click chip B → `HIT id=btn-b`; not btn-a |
| `top/click-btn-a` | needs-pty, slow | Top anchor: click A still resolves btn-a |
| `bottom/click-btn-a` | needs-pty, slow | Bottom anchor: click A still resolves btn-a |

## Harness phases (Run)

| Phase | Action |
|-------|--------|
| 0 | Build fixture; resolve tty-watch **binary** (PATH / build sibling) |
| 1 | Per-leaf temp Home; unique session id |
| 2 | `exec tty-watch run --detach --session-id … -- fixture --anchor=…` |
| 3 | Poll `tty-watch snapshot` for paint / `ORIGIN=` |
| 4 | If query-cursor: `tty-watch send … --query-cursor --json` |
| 5 | If click: parse ORIGIN; `tty-watch send … --click --row --col` |
| 6 | Poll snapshot for `HIT=` (click leaves) |
| 7 | `tty-watch kill` (always, best-effort) |

## How to Run

```sh
cd external/dot-pkgs-master-2026-07-18-1/go-pkgs

# pure (must stay GREEN)
doctest test ./tui/mouse/tests/hit-test/...

# headless (needs tty-watch binary on PATH or sibling source tree)
doctest vet ./tui/mouse/tests/headless
doctest test --label 'needs-pty || slow' ./tui/mouse/tests/headless/...
```

```go
import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

// Request configures one headless geometry session.
// Anchor: "top" | "mid" | "bottom"
// Action: "query-cursor" | "click"
type Request struct {
	Anchor string
	Action string

	// Click target in view-local coords (ignored for query-cursor).
	LocalY    int
	ClickCol  int // 0-based absolute X; default 5
	WantHitID string

	// Filled by root Setup / Run.
	Home        string
	SessionID   string
	FixtureBin  string
	TTYWatchBin string
}

type Response struct {
	SessionID string

	// Snapshot text after paint wait / post-click wait.
	Snapshot string

	// Parsed ORIGIN line (if any).
	OriginY   int
	ViewLines int
	HasOrigin bool

	// Query-cursor JSON (0-based).
	CursorRow int
	CursorCol int
	HasCursor bool

	// Parsed last HIT line.
	HitID     string
	HitLocalY int
	HitKind   string
	HasHit    bool

	// Raw HIT lines found in snapshot.
	HitLines []string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}
	if req.FixtureBin == "" || req.TTYWatchBin == "" {
		return nil, fmt.Errorf("missing binaries: fixture=%q tty-watch=%q", req.FixtureBin, req.TTYWatchBin)
	}
	if req.Home == "" {
		return nil, fmt.Errorf("missing Home")
	}
	if req.SessionID == "" {
		req.SessionID = "mouse-hl-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	if req.ClickCol <= 0 {
		req.ClickCol = 5
	}
	if req.Anchor == "" {
		req.Anchor = "mid"
	}

	resp := &Response{SessionID: req.SessionID}
	defer func() {
		_, _ = runTTYWatch(req, "kill", req.SessionID)
	}()

	// Phase 2: detach via tty-watch binary (no Go import of tty-watch).
	if err := runDetach(req); err != nil {
		return resp, fmt.Errorf("run detach: %w", err)
	}

	// Phase 3: wait paint / ORIGIN
	snap, err := waitSnapshot(req, 8*time.Second, func(s string) bool {
		return strings.Contains(s, "ORIGIN=") ||
			strings.Contains(s, "btn-a") ||
			strings.Contains(s, "fixture-inline")
	})
	resp.Snapshot = snap
	if err != nil {
		return resp, fmt.Errorf("paint wait: %w", err)
	}
	parseOriginInto(resp, snap)

	if req.Action == "query-cursor" {
		row, col, qerr := sendQueryCursor(req)
		if qerr != nil {
			return resp, fmt.Errorf("query-cursor: %w", qerr)
		}
		resp.CursorRow, resp.CursorCol, resp.HasCursor = row, col, true
		if !resp.HasOrigin {
			if s2, e2 := snapshotOnce(req); e2 == nil {
				resp.Snapshot = s2
				parseOriginInto(resp, s2)
			}
		}
		return resp, nil
	}

	if req.Action == "click" {
		if !resp.HasOrigin {
			snap2, e2 := waitSnapshot(req, 4*time.Second, func(s string) bool {
				return strings.Contains(s, "ORIGIN=")
			})
			if e2 == nil {
				resp.Snapshot = snap2
				parseOriginInto(resp, snap2)
			}
		}
		if !resp.HasOrigin {
			return resp, fmt.Errorf("no ORIGIN= line before click (snapshot:\n%s)", trimSnap(resp.Snapshot))
		}
		absY := resp.OriginY + req.LocalY
		absX := req.ClickCol
		if err := sendClick(req, absY, absX); err != nil {
			return resp, fmt.Errorf("click: %w", err)
		}
		wantPrefix := "HIT id="
		if req.WantHitID != "" {
			wantPrefix = "HIT id=" + req.WantHitID
		}
		snap3, e3 := waitSnapshot(req, 6*time.Second, func(s string) bool {
			return strings.Contains(s, wantPrefix) || strings.Contains(s, "HIT id=")
		})
		resp.Snapshot = snap3
		parseHitsInto(resp, snap3)
		if e3 != nil {
			return resp, fmt.Errorf("HIT wait: %w", e3)
		}
		return resp, nil
	}

	return resp, fmt.Errorf("unknown Action %q", req.Action)
}

var (
	originRe = regexp.MustCompile(`ORIGIN=(\d+)\s+VIEW=(\d+)`)
	hitRe    = regexp.MustCompile(`HIT id=([^\s]+)\s+localY=(\d+)\s+kind=([^\s]+)`)
)

func parseOriginInto(resp *Response, snap string) {
	m := originRe.FindStringSubmatch(snap)
	if m == nil {
		return
	}
	oy, _ := strconv.Atoi(m[1])
	vl, _ := strconv.Atoi(m[2])
	resp.OriginY, resp.ViewLines, resp.HasOrigin = oy, vl, true
}

func parseHitsInto(resp *Response, snap string) {
	all := hitRe.FindAllStringSubmatch(snap, -1)
	resp.HitLines = nil
	for _, m := range all {
		resp.HitLines = append(resp.HitLines, m[0])
	}
	if len(all) == 0 {
		return
	}
	last := all[len(all)-1]
	resp.HitID = last[1]
	resp.HitLocalY, _ = strconv.Atoi(last[2])
	resp.HitKind = last[3]
	resp.HasHit = true
}

func trimSnap(s string) string {
	if len(s) > 800 {
		return s[:800] + "…"
	}
	return s
}

// ttyWatchEnv sets TTY_WATCH_HOME for a host command. Clears ambient registry
// subdir so Home/registry is consistent for parent and serve child.
func ttyWatchEnv(home string) []string {
	env := os.Environ()
	out := make([]string, 0, len(env)+2)
	for _, e := range env {
		if strings.HasPrefix(e, "TTY_WATCH_HOME=") ||
			strings.HasPrefix(e, "TTY_WATCH_REGISTRY_SUBDIR=") {
			continue
		}
		out = append(out, e)
	}
	out = append(out, "TTY_WATCH_HOME="+home)
	return out
}

func runTTYWatch(req *Request, args ...string) (string, error) {
	cmd := exec.Command(req.TTYWatchBin, args...)
	cmd.Env = ttyWatchEnv(req.Home)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.String(), err
}

func runDetach(req *Request) error {
	// tty-watch run --detach --session-id SID -- fixture --anchor=…
	args := []string{
		"run", "--detach", "--session-id", req.SessionID, "--",
		req.FixtureBin, "--anchor=" + req.Anchor,
	}
	out, err := runTTYWatch(req, args...)
	if err != nil {
		return fmt.Errorf("%w\n%s", err, strings.TrimSpace(out))
	}
	return nil
}

func snapshotOnce(req *Request) (string, error) {
	return runTTYWatch(req, "snapshot", req.SessionID)
}

func waitSnapshot(req *Request, timeout time.Duration, ok func(string) bool) (string, error) {
	deadline := time.Now().Add(timeout)
	var last string
	var lastErr error
	for time.Now().Before(deadline) {
		s, err := snapshotOnce(req)
		last, lastErr = s, err
		if err == nil && ok(s) {
			return s, nil
		}
		time.Sleep(150 * time.Millisecond)
	}
	if lastErr != nil {
		return last, fmt.Errorf("timeout after %s: last err: %v\n%s", timeout, lastErr, trimSnap(last))
	}
	return last, fmt.Errorf("timeout after %s waiting for marker\n%s", timeout, trimSnap(last))
}

func sendQueryCursor(req *Request) (row, col int, err error) {
	out, err := runTTYWatch(req, "send", req.SessionID, "--query-cursor", "--json")
	if err != nil {
		return 0, 0, fmt.Errorf("%w\n%s", err, out)
	}
	// stdout may include trailing newline only
	line := strings.TrimSpace(out)
	var parsed struct {
		Row int `json:"row"`
		Col int `json:"col"`
	}
	if jerr := json.Unmarshal([]byte(line), &parsed); jerr != nil {
		return 0, 0, fmt.Errorf("parse query-cursor json: %w (out=%q)", jerr, out)
	}
	return parsed.Row, parsed.Col, nil
}

func sendClick(req *Request, row, col int) error {
	out, err := runTTYWatch(req,
		"send", req.SessionID,
		"--click",
		"--row", strconv.Itoa(row),
		"--col", strconv.Itoa(col),
		"--json",
	)
	if err != nil {
		return fmt.Errorf("%w\n%s", err, out)
	}
	return nil
}
```
