# ptywrap Live Screen Model — Snapshot Fidelity

Contract tests for a **persistent per-session VT (cell model)** so
`attach_mode=snapshot` exports the live screen cells instead of cold-replaying
a truncated scrollback ring. Fixes sticky chrome/footer loss for dirty-region
TUIs (e.g. Grok): early paint of prompt/footer stays on the live screen even
after dirty-only updates push those bytes out of the 256 KiB ring.

# DSN (Domain Specific Notion)

**Participants**

- **Session (`ptywrap`)** — owns long-lived PTY child, scrollback ring, and
  (after fix) a **live `screen` VT** sized to session `cols×rows`.
- **PTY child (fixture TUI)** — programmatic ANSI process that paints sticky
  chrome once (bottom prompt/footer), then emits many dirty-region-only frames
  (`?2026h` / `?2026l` + CUP to top/mid rows) without repainting sticky lines.
- **Scrollback ring** — existing byte log, still trimmed at `maxScrollback`
  (256 KiB). Secondary history only; **not** source of truth for “what is on
  screen now.”
- **Live screen VT** — long-lived cell model: every PTY output chunk is applied
  before (or as) scrollback append; resize updates geometry; snapshot **exports
  cells** from this model.
- **Snapshot attach** — `attach_mode=snapshot`: one-shot frame, `roleSnapshot`,
  does not claim writer; close must not kill child.
- **Test harness** — httptest + `RegisterAPI`; runs fixture TUI; optional
  resize; optional ring pressure; collects snapshot `WSOutput` + liveness.

**Behaviors**

- After sticky paint + dirty-only updates, snapshot plain text contains sticky
  markers (`STICKY_FOOTER`, `STICKY_PROMPT`).
- After sticky paint + ≥ `maxScrollback` dirty-only bytes (sticky sequences
  left the ring), snapshot **still** contains sticky markers (RED on cold
  scrollback replay; GREEN with live VT export).
- ≥3 snapshot attaches after sticky+dirty leave child alive and session listed.
- After resize to known geometry, sticky paint + dirty updates, snapshot still
  contains sticky; frame remains a rendered CUP-style snapshot (not raw ring).

## Version

0.0.2

## Decision Tree

```
go-pkgs/shell/ptywrap/tests/live-screen-model/
├── DOCTEST.md
├── SETUP.md
├── testdata/
│   └── sticky_dirty_tui.py          # ANSI fixture: sticky once, then dirty frames
├── sticky-after-dirty/              # fidelity under dirty-region updates
├── sticky-after-scrollback-pressure/# RED-defining: ring trim + sticky retained
├── multi-snapshot-keeps-child/      # lifecycle regression after sticky+dirty
└── resize-then-sticky/              # resize geometry then sticky still present
```

Parameter ranking (most → least significant):

1. **Screen-stress mode** — dirty-only / ring-pressure / resize / multi-attach
   (what the live model must survive before snapshot export)
2. **Attach mode** — always `snapshot` for fidelity/lifecycle leaves
3. **Geometry** — default 80×24 vs explicit resize (e.g. 100×30)
4. **Markers** — `STICKY_FOOTER` / `STICKY_PROMPT` / latest `DIRTY_*`

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `sticky-after-dirty` | Sticky chrome survives dirty-only updates (no ring pressure) |
| 2 | `sticky-after-scrollback-pressure` | Sticky survives ≥ maxScrollback dirty-only bytes (RED-defining) |
| 3 | `multi-snapshot-keeps-child` | After sticky+dirty, ≥3 snapshots; child still alive |
| 4 | `resize-then-sticky` | Resize then sticky+dirty; snapshot still has sticky |

## How to Run

```sh
# from go-pkgs module root
doctest vet ./shell/ptywrap/tests/live-screen-model
doctest test ./shell/ptywrap/tests/live-screen-model/...
```

Related tree (non-destructive snapshot contract without live-VT stress):
`shell/ptywrap/tests/snapshot-attach/`.

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

// Request configures a live-screen-model harness phase.
type Request struct {
	// Phase selects the scenario runner in Run.
	//   live-screen-sticky-after-dirty
	//   live-screen-sticky-after-scrollback-pressure
	//   live-screen-multi-snapshot-keeps-child
	//   live-screen-resize-then-sticky
	Phase string

	ServerBase string

	AttachMode string // default "snapshot"

	// StickyMarker / PromptMarker are expected substrings in snapshot text.
	StickyMarker string // default STICKY_FOOTER
	PromptMarker string // default STICKY_PROMPT

	// DirtyIters is how many dirty-only frames to emit when not pressure-driven.
	DirtyIters int

	// PressureBytes, when > 0, drives dirty emission until at least this many
	// bytes of dirty frames have been written (used to exceed maxScrollback).
	PressureBytes int

	// ResizeCols / ResizeRows, when both > 0, writer-resizes before paint wait.
	ResizeCols int
	ResizeRows int

	// RepeatCount is N snapshot attaches for multi-snapshot phases (default 3).
	RepeatCount int

	// ExpectDirty is true when Assert should require a DIRTY_ substring.
	ExpectDirty bool
}

// Response holds snapshot and liveness observations.
type Response struct {
	SessionID     string
	WSOutput      string
	SnapshotCount int

	ProcessAlive  bool
	SessionListed bool
	SessionCount  int

	PTYCols int
	PTYRows int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	if req.ServerBase == "" {
		return nil, fmt.Errorf("ServerBase not set; root Setup must start server")
	}
	mode := strings.TrimSpace(req.AttachMode)
	if mode == "" {
		mode = "snapshot"
	}
	req.AttachMode = mode

	if req.StickyMarker == "" {
		req.StickyMarker = "STICKY_FOOTER"
	}
	if req.PromptMarker == "" {
		req.PromptMarker = "STICKY_PROMPT"
	}

	switch req.Phase {
	case "live-screen-sticky-after-dirty":
		return runStickyScenario(t, req, stickyScenarioOpts{
			dirtyIters:     defaultPositive(req.DirtyIters, 30),
			pressureBytes:  0,
			repeatSnapshot: 1,
			waitResize:     false,
		})
	case "live-screen-sticky-after-scrollback-pressure":
		// Production maxScrollback is 256 KiB; emit past the ring so early
		// sticky paint is no longer present in scrollback for cold replay.
		pressure := req.PressureBytes
		if pressure <= 0 {
			pressure = 256*1024 + 64*1024 // 320 KiB dirty payload
		}
		return runStickyScenario(t, req, stickyScenarioOpts{
			dirtyIters:     0,
			pressureBytes:  pressure,
			repeatSnapshot: 1,
			waitResize:     false,
		})
	case "live-screen-multi-snapshot-keeps-child":
		n := defaultPositive(req.RepeatCount, 3)
		return runStickyScenario(t, req, stickyScenarioOpts{
			dirtyIters:     defaultPositive(req.DirtyIters, 20),
			pressureBytes:  0,
			repeatSnapshot: n,
			waitResize:     false,
		})
	case "live-screen-resize-then-sticky":
		cols := req.ResizeCols
		rows := req.ResizeRows
		if cols <= 0 {
			cols = 100
		}
		if rows <= 0 {
			rows = 30
		}
		req.ResizeCols, req.ResizeRows = cols, rows
		return runStickyScenario(t, req, stickyScenarioOpts{
			dirtyIters:     defaultPositive(req.DirtyIters, 20),
			pressureBytes:  0,
			repeatSnapshot: 1,
			waitResize:     true,
			resizeCols:     cols,
			resizeRows:     rows,
		})
	default:
		return nil, fmt.Errorf("unknown phase %q", req.Phase)
	}
}

func defaultPositive(v, def int) int {
	if v <= 0 {
		return def
	}
	return v
}

type stickyScenarioOpts struct {
	dirtyIters     int
	pressureBytes  int
	repeatSnapshot int
	waitResize     bool
	resizeCols     int
	resizeRows     int
}

// runStickyScenario starts the ANSI fixture TUI, optionally resizes, waits for
// sticky (+ optional dirty completion marker), takes N snapshot attaches, and
// reports WSOutput / ProcessAlive / SessionListed.
func runStickyScenario(t *testing.T, req *Request, opts stickyScenarioOpts) (*Response, error) {
	t.Helper()
	resp := &Response{}

	script := fixtureScriptPath()
	if _, err := os.Stat(script); err != nil {
		return nil, fmt.Errorf("fixture script: %w", err)
	}

	token := fmt.Sprintf("lsm-%d-%d", os.Getpid(), time.Now().UnixNano())
	args := []string{
		script,
		"--token", token,
		"--sticky", req.StickyMarker,
		"--prompt", req.PromptMarker,
	}
	if opts.pressureBytes > 0 {
		args = append(args, "--pressure-bytes", fmt.Sprintf("%d", opts.pressureBytes))
	} else {
		args = append(args, "--dirty-iters", fmt.Sprintf("%d", opts.dirtyIters))
	}
	if opts.waitResize {
		args = append(args,
			"--wait-cols", fmt.Sprintf("%d", opts.resizeCols),
			"--wait-rows", fmt.Sprintf("%d", opts.resizeRows),
			"--wait-timeout-ms", "8000",
		)
	}

	cmd := append([]string{"python3"}, args...)
	id, err := createSessionREST(t, req.ServerBase, cmd, "", token)
	if err != nil {
		return nil, err
	}
	resp.SessionID = id

	pid, err := findPIDByToken(t, token)
	if err != nil {
		_ = deleteSessionREST(req.ServerBase, id)
		return nil, fmt.Errorf("session %s: resolve child: %w", id, err)
	}
	t.Cleanup(func() {
		if processAlive(pid) {
			_ = killPID(pid)
		}
		_ = deleteSessionREST(req.ServerBase, id)
	})

	if opts.waitResize {
		if err := wsResizeOnly(t, req.ServerBase, id, opts.resizeCols, opts.resizeRows); err != nil {
			return nil, fmt.Errorf("resize: %w", err)
		}
		resp.PTYCols = opts.resizeCols
		resp.PTYRows = opts.resizeRows
	}

	// Wait until the fixture has finished its dirty loop (DONE file / DIRTY_DONE)
	// so pressure scenarios have actually advanced the ring. Then take the
	// authoritative snapshot used for assertions.
	deadline := time.Now().Add(45 * time.Second)
	var lastOut string
	for time.Now().Before(deadline) {
		if fixtureDoneFile(token) {
			break
		}
		// Opportunistic snapshot while waiting (also exercises multi-frame path).
		out, snapErr := wsAttachSnapshot(t, req.ServerBase, id, req.AttachMode, 400*time.Millisecond)
		if snapErr == nil && strings.TrimSpace(out) != "" {
			lastOut = out
			if strings.Contains(out, "DIRTY_DONE") {
				break
			}
			// Non-pressure: sticky + at least one dirty cell is enough once DONE
			// is slow; still prefer DONE when present.
			if opts.pressureBytes == 0 &&
				strings.Contains(out, req.StickyMarker) &&
				strings.Contains(out, "DIRTY_") {
				// Keep waiting a bit for DONE so multi-snapshot leaves see a stable child,
				// but do not require DONE if the marker pair is already visible.
				if fixtureDoneFile(token) {
					break
				}
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	// Final snapshot after wait (post-pressure when DONE exists).
	out, snapErr := wsAttachSnapshot(t, req.ServerBase, id, req.AttachMode, 2*time.Second)
	if snapErr != nil {
		if lastOut == "" {
			return nil, fmt.Errorf("snapshot wait: %w", snapErr)
		}
	} else if strings.TrimSpace(out) != "" {
		lastOut = out
	}
	if strings.TrimSpace(lastOut) == "" {
		return nil, fmt.Errorf("snapshot wait: empty frame after fixture run (session=%s)", id)
	}

	n := opts.repeatSnapshot
	if n <= 0 {
		n = 1
	}
	resp.WSOutput = lastOut
	if strings.TrimSpace(lastOut) != "" {
		resp.SnapshotCount = 1
	}
	for i := resp.SnapshotCount; i < n; i++ {
		out, snapErr := wsAttachSnapshot(t, req.ServerBase, id, req.AttachMode, 1500*time.Millisecond)
		if snapErr != nil {
			return nil, fmt.Errorf("snapshot %d/%d: %w", i+1, n, snapErr)
		}
		if strings.TrimSpace(out) != "" {
			resp.SnapshotCount++
			resp.WSOutput = out
		}
		time.Sleep(50 * time.Millisecond)
	}

	time.Sleep(200 * time.Millisecond)
	sessions, err := listSessions(t, req.ServerBase)
	if err != nil {
		return nil, err
	}
	resp.ProcessAlive = processAlive(pid)
	resp.SessionListed = sessionInList(sessions, id)
	resp.SessionCount = len(sessions)
	return resp, nil
}

func fixtureScriptPath(d *session.Doctest) string {
	// Leaf packages run with working directory = leaf dir; d.DOCTEST_ROOT is tree root.
	return filepath.Join(d.DOCTEST_ROOT, "testdata", "sticky_dirty_tui.py")
}

func fixtureDoneFile(token string) bool {
	// Fixture writes DONE file under TMPDIR when pressure/dirty loop finishes.
	p := filepath.Join(os.TempDir(), "ptywrap-lsm-done-"+token)
	_, err := os.Stat(p)
	return err == nil
}
```
