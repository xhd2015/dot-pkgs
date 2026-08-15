# tui/mouse — Pure hit-test, origin, demux, tracker

## Version
0.0.2

Coverage-backfill doctests for pure APIs in
`github.com/xhd2015/dot-pkgs/go-pkgs/tui/mouse`. Implementation already exists;
leaves seal contracts for HitTest, Resolve (known + dual), OriginFromCPR
(accept / reject), DemuxCPR (CPR peel + SGR mouse forward), and light Tracker
state transitions. Out of scope for **this pure root**: headless/tty-watch (see nested
`tests/headless/DOCTEST.md` P2 tree), wrk layout, Filter I/O (package unit tests).

## DSN (Domain Specific Notion)

### Participants

- **App / TUI** — paints a multi-line inline (non-alt-screen) UI frame and
  registers rectangular hit regions with action IDs (e.g. Run chips).
- **Terminal** — reports absolute mouse coordinates and replies to CSI 6n with
  CPR (`ESC [ row ; col R`, 1-based). Also emits SGR mouse (`ESC [ < … M/m`).
- **HitTest** — maps view-local `(x, localY)` onto the first matching half-open
  rectangle (`y0 ≤ y < y1`; optional `x0 ≤ x < x1` when `x1 > x0`).
- **Resolve** — maps absolute mouse `(absX, absY)` to a Hit using either a
  **known** origin (`localY = absY - originY`) or **dual-origin** (try top
  `localY=absY` then bottom `localY=absY-(height-viewLines)`; prefer non-empty
  IDs when candidates disagree).
- **OriginFromCPR** — derives 0-based origin when cursor sits on the last line
  of a `viewLines`-tall frame: `originY0 = row1 - viewLines`. Live rule rejects
  `row1 < viewLines` (stale probe must not look top-anchored).
- **DemuxCPR** — peels complete CPR sequences from a byte stream; forwards mouse
  SGR and all other bytes; returns incomplete trailing hold.
- **Tracker** — owns CSI 6n measurement phase: Unknown → Pending (FrameSuffix
  embeds query) → Known (good CPR) or Failed (bad CPR); OnResize invalidates
  back to Unknown.

### Behaviors

- HitTest: first match wins; `localY == y1` misses (half-open).
- Resolve known mid-pane: `originY=6`, click abs rows 9/10 → add-changes /
  gen-commit-msg with `Kind=known`.
- Resolve dual-top: top-anchored click on gen row is gen-commit-msg, not
  tag-next.
- Resolve dual-bottom: bottom-anchored relative click still resolves gen.
- OriginFromCPR: `row1=26, viewLines=20` → origin 6 ok; `row1=9` → !ok.
- DemuxCPR: CPR then SGR mouse → one event, mouse bytes forwarded, empty rest.
- Tracker: mid-pane CPR → Known origin 6; row1 < viewLines → Failed; resize →
  Unknown and re-emit FrameSuffix.

## Decision Tree

```
mouse pure
├── hit-test/                        view-local HitTest
│   ├── hit-run-chip/                (x,y) inside run rectangle → ID run
│   └── y1-half-open-miss/           localY == y1 → miss
├── resolve/                         absolute → Hit
│   ├── known-mid/                   OriginY set (mid-pane 6)
│   │   ├── absy-9-add-changes/      absY=9 → add-changes, LocalY=3
│   │   └── absy-10-gen-commit-msg/  absY=10 → gen-commit-msg, LocalY=4
│   ├── dual-top/                    OriginY nil; top-anchored gen click
│   └── dual-bottom/                 OriginY nil; bottom-anchored gen click
├── origin-from-cpr/                 live origin derivation
│   ├── valid/                       row1=26 view=20 → origin 6
│   └── reject-row-lt-view/          row1=9 view=20 → !ok
├── demux-cpr-mouse/                 CPR peel + SGR mouse forward
└── tracker/                         light measurement state machine
    ├── mid-pane-known/              FrameSuffix + good CPR → Known
    ├── bad-cpr-failed/              row1 < viewLines → Failed
    └── resize-unknown/              Known then OnResize → Unknown
```

## Test Index

| Leaf | Description |
|------|-------------|
| `hit-test/hit-run-chip` | HitTest hits run chip at (65,3) |
| `hit-test/y1-half-open-miss` | HitTest misses at localY == y1 |
| `resolve/known-mid/absy-9-add-changes` | Known origin: absY=9 → add-changes |
| `resolve/known-mid/absy-10-gen-commit-msg` | Known origin: absY=10 → gen-commit-msg |
| `resolve/dual-top` | Dual: top absY=4 → gen-commit-msg not tag-next |
| `resolve/dual-bottom` | Dual: bottom-relative gen still gen (Kind=bottom) |
| `origin-from-cpr/valid` | OriginFromCPR(26,20) → 6, ok |
| `origin-from-cpr/reject-row-lt-view` | OriginFromCPR(9,20) → !ok |
| `demux-cpr-mouse` | DemuxCPR peels CPR, forwards SGR mouse |
| `tracker/mid-pane-known` | Tracker OnCPR mid-pane → PhaseKnown origin 6 |
| `tracker/bad-cpr-failed` | Tracker bad CPR → PhaseFailed |
| `tracker/resize-unknown` | Tracker OnResize → PhaseUnknown + re-probe |

## Unit expansion matrix (implementer)

Expand `mouse_test.go` (or split `*_test.go`) for table-driven edges not sealed
as doctest leaves — do not replace existing tests:

| Edge | Notes |
|------|--------|
| `BottomOriginY` | height/viewLines ≤0, clamp when viewLines > height |
| `OriginFromCPRClamped` | row1 < viewLines clamps to 0 (legacy) |
| `ParseCPR` | incomplete, malformed, multi-CPR first-wins |
| `Resolve` miss | known/dual no hit → OK=false LocalY=-1 |
| `Tracker.FrameSuffix` | once per Unknown; empty while Pending/Known |
| PreferID / empty-ID dual | optional if product uses PreferID field |

## How to Run

```sh
cd external/dot-pkgs-master-2026-07-18-1/go-pkgs
doctest vet ./tui/mouse/tests
doctest test ./tui/mouse/tests/hit-test/...
doctest test ./tui/mouse/tests/...   # pure green; nested headless is separate root (P2 RED until fixture)
doctest vet ./tui/mouse/tests/headless
doctest test ./tui/mouse/tests/headless/...
go test ./tui/mouse/ -count=1
```

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/tui/mouse"
)

// Op selects which pure mouse API Run exercises.
// Values: "hit-test" | "resolve" | "origin-from-cpr" | "demux" | "tracker".
type Request struct {
	Op string

	// HitTest
	Hits          []mouse.Hit
	X, LocalY     int

	// Resolve (also uses Hits)
	AbsX, AbsY    int
	Height        int
	ViewLines     int
	OriginY       *int

	// OriginFromCPR (Row1 + ViewLines)
	Row1 int

	// DemuxCPR
	DemuxHold []byte
	DemuxData []byte

	// Tracker: ordered pure steps after NewTracker.
	// Kinds: "frame-suffix" | "on-cpr" | "on-resize" | "on-view-lines"
	TrackerSteps []TrackerStep
}

type TrackerStep struct {
	Kind              string
	Height, ViewLines int
	Row1, Col1        int
}

type Response struct {
	// HitTest
	Hit   mouse.Hit
	HitOK bool

	// Resolve
	Resolve mouse.ResolveResult

	// OriginFromCPR
	OriginY0 int
	OriginOK bool

	// DemuxCPR
	Events  []mouse.CPR
	Forward []byte
	Rest    []byte

	// Tracker final snapshot
	Phase          mouse.Phase
	TrackerOriginY *int
	LastSuffix     string
	LastOnCPR      bool
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	resp := &Response{}
	switch req.Op {
	case "hit-test":
		h, ok := mouse.HitTest(req.Hits, req.X, req.LocalY)
		resp.Hit = h
		resp.HitOK = ok
	case "resolve":
		resp.Resolve = mouse.Resolve(mouse.ResolveOpts{
			AbsX: req.AbsX, AbsY: req.AbsY,
			Height: req.Height, ViewLines: req.ViewLines,
			OriginY: req.OriginY, Hits: req.Hits,
		})
	case "origin-from-cpr":
		oy, ok := mouse.OriginFromCPR(req.Row1, req.ViewLines)
		resp.OriginY0 = oy
		resp.OriginOK = ok
	case "demux":
		ev, fwd, rest := mouse.DemuxCPR(req.DemuxHold, req.DemuxData)
		resp.Events = ev
		resp.Forward = fwd
		resp.Rest = rest
	case "tracker":
		tr := mouse.NewTracker()
		for _, step := range req.TrackerSteps {
			switch step.Kind {
			case "frame-suffix":
				resp.LastSuffix = tr.FrameSuffix(step.Height, step.ViewLines)
			case "on-cpr":
				resp.LastOnCPR = tr.OnCPR(step.Row1, step.Col1)
			case "on-resize":
				tr.OnResize()
			case "on-view-lines":
				tr.OnViewLines(step.ViewLines)
			default:
				return nil, fmt.Errorf("unknown tracker step kind %q", step.Kind)
			}
		}
		resp.Phase = tr.Phase()
		resp.TrackerOriginY = tr.OriginY()
	default:
		return nil, fmt.Errorf("unknown Op %q", req.Op)
	}
	return resp, nil
}
```
