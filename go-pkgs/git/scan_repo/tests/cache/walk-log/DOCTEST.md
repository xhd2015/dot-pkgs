# scan_repo — walk JSONL cold + consume (gen_end G+1) + adaptive budget (P3/P4)

## Version
0.0.2

Nested doc tests for **Phase 3** (cold walk log) and **Phase 4** (incremental
consume of `walk.jsonl` from the durable cursor, seal `gen_end` G+1, adaptive
sync budget from time since `last_scan_end`).

Cold seal + warm consume + adaptive budget are wired in Scan / walk_log helpers
(implemented; leaves expect green). Consume isolates unit warm-refresh via
`WarmRefreshBudget=-1` so observations focus on walk-log work.

**Out of scope:** wrk CLI filter / two-base (see wrk `scan-git-repos`); multi-call
cold resume; side-channel async consume details beyond “0 sync work when
delta &lt; 10s”.

Nested `DOCTEST.md` isolates `Request`/`Response`/`Run` from the parent mirror
`CacheOp` contract and from pure index I/O trees.

## DSN (Domain Specific Notion)

### Participants

- **Caller** — supplies one scan root, explicit temp `CacheRoot`, `NoCache`,
  and for P4: optional `LastScanEnd`, injectable `Now`, and post-cold mutations.
- **Scan (cold)** — full live walk when cache is enabled and the root is not
  warm-eligible; discovers mains; appends each walked directory as a **visit**
  event; on success seals **`gen_end` gen=1** and sets the walk cursor to EOF.
- **Scan (warm + consume)** — second (or later) Scan with cache when the root
  is warm-eligible: serves index; **also** consumes `walk.jsonl` from the
  durable cursor under a **sync budget** selected from time since last scan end.
- **Walk log store** — append-only JSONL at `<CacheRoot>/home/walk.jsonl`.
  Each line is one JSON object with an `op` field.
- **Visit event** — `{"op":"visit","path":"<abs-dir>"}` (extra fields allowed).
- **Gone event** — `{"op":"gone","path":"<abs-dir>"}` when a prior visit path
  no longer exists on re-list (extra fields allowed).
- **Generation seal** — `{"op":"gen_end","gen":N}` ends generation N. First
  cold uses **1**. When the consumer processes `gen_end` G, it appends
  `gen_end` **G+1** after finishing that generation’s consume cycle.
- **Walk cursor** — `<CacheRoot>/home/walk.cursor.json` with byte **offset**
  into `walk.jsonl`. After cold seal: offset = sealed EOF. After consume:
  offset advances as events are processed and ends at the new sealed EOF when
  gen_end G+1 is written.
- **Last scan end / meta** — durable `last_scan_end` (RFC3339 or unix) under
  `<CacheRoot>/home/meta.json` (or equivalent). Used when
  `Options.LastScanEnd` is zero. Updated when a Scan finishes (product detail).
- **Budget selector** — given `delta = Now() - last_scan_end`:
  - **delta &lt; 10s** → side / best-effort: **0 sync** re-list budget in tests
    (no guaranteed cursor progress / no new discover from consume).
  - **10s ≤ delta &lt; 60s** → **500ms** sync re-list budget.
  - **delta ≥ 60s** → **1s** sync re-list budget.
- **Options fields required for P4 (document for implementer):**
  - `LastScanEnd time.Time` — optional override of meta `last_scan_end`;
    zero means read meta (or treat as “ancient” / full budget if missing —
    product documents; tests inject explicit times).
  - `Now func() time.Time` — already on Options; nil → `time.Now`.
  - (Existing) `WarmRefreshBudget` — consume leaves set **negative** to
    **disable** unit warm-refresh so observations isolate walk-log consume.
- **NoCache path** — no walk log / cursor under `home/`.
- **Filesystem** — fake `.git` fixtures; no enrichment, no git CLI.

### Behaviors

- **Cold writes visits + seals gen 1:** first successful Scan with cache
  produces `home/walk.jsonl` visits and ends with `gen_end` gen=1; cursor at
  log EOF.
- **Consume seals gen 2:** after gen_end 1 is present and the cursor sits at
  that seal, a second Scan with a non-zero sync budget (e.g. delta ≥ 60s)
  processes prior visit lines (re-list), may append new events after gen_end 1,
  and when gen_end 1 is consumed appends **gen_end 2**; cursor moves to the
  new EOF.
- **Cursor advances:** after successful consume that writes gen_end 2 (or any
  post-seal append), `walk.cursor.json` offset is greater than the cold seal
  offset and equals the new log size.
- **Gone:** if a directory that had a visit is removed before the second Scan,
  consume appends a `gone` event for that path (when budget allows re-list).
- **Budget tiers:** pure selection via exported
  `WalkConsumeSyncBudget(sinceLast time.Duration) time.Duration` (or
  `TestExported_WalkConsumeSyncBudget` if kept unexported). Behavioral
  delta-lt-10s: with WarmRefreshBudget disabled and delta &lt; 10s, a new repo
  planted after cold is **not** discovered via walk-consume sync work.
- **NoCache skips walk log:** unchanged from P3.

## Decision Tree

```
walk-log                         [nested — walk JSONL cold + consume + budget]
├── cold/                        [NoCache=false — single successful cold Scan]
│   └── complete/                # visits + gen_end gen=1 + cursor at EOF
├── no-cache/                    [NoCache=true]
│   └── skips-write/             # no walk.jsonl / walk.cursor.json under home/
└── consume/                     [P4: cold seed then second Scan; isolate warm unit refresh]
    ├── seal-gen-end-2/          # processes gen_end 1 → append gen_end 2
    ├── cursor-advance/          # cursor offset > cold EOF; equals new log size
    ├── gone/                    # deleted visited dir → gone event after gen_end 1
    └── budget/                  [split on Now − LastScanEnd]
        ├── delta-lt-10s/        # 0 sync budget; no new consume discover
        ├── delta-10s-to-60s/    # WalkConsumeSyncBudget → 500ms
        └── delta-ge-60s/        # WalkConsumeSyncBudget → 1s (+ full seal ok)
```

## Test Index

| Leaf | Mode | Description |
|------|------|-------------|
| `cold/complete` | cold | Cold Scan writes visits + seals gen_end 1; cursor at EOF |
| `no-cache/skips-write` | no-cache | NoCache leaves no walk log artifacts under home/ |
| `consume/seal-gen-end-2` | consume | Second Scan seals gen_end 2 after processing gen_end 1 |
| `consume/cursor-advance` | consume | Cursor advances past cold EOF to new sealed size |
| `consume/gone` | consume | Removed visit path yields `gone` event on re-list |
| `consume/budget/delta-lt-10s` | budget | delta &lt; 10s → 0 sync; new post-cold repo not found via consume |
| `consume/budget/delta-10s-to-60s` | budget | 10s ≤ delta &lt; 60s → SelectedBudget 500ms |
| `consume/budget/delta-ge-60s` | budget | delta ≥ 60s → SelectedBudget 1s |

## How to Run

```sh
doctest vet ./go-pkgs/git/scan_repo/tests/cache/walk-log/
doctest test -v ./go-pkgs/git/scan_repo/tests/cache/walk-log/
```

From monorepo / worktree with nested external:

```sh
doctest vet ./external/dot-pkgs-master-2026-07-15/go-pkgs/git/scan_repo/tests/cache/walk-log/
doctest test -v ./external/dot-pkgs-master-2026-07-15/go-pkgs/git/scan_repo/tests/cache/walk-log/
```

P4 only:

```sh
doctest test -v ./external/dot-pkgs-master-2026-07-15/go-pkgs/git/scan_repo/tests/cache/walk-log/consume/
```

```go
import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

// Library universe for single-root fixtures (same as P2 index serve).
const walkUniverseHome = "home"

// WalkEvent is one JSONL object from walk.jsonl.
// Product may add fields; tests require op, path for visit/gone, gen for gen_end.
type WalkEvent struct {
	Op   string `json:"op"`
	Path string `json:"path,omitempty"`
	Gen  int    `json:"gen,omitempty"`
}

// WalkCursor is the durable cursor document at walk.cursor.json.
type WalkCursor struct {
	Offset int64 `json:"offset"`
}

type Request struct {
	Roots     []string
	CacheRoot string
	NoCache   bool
	Refresh   bool
	Debug     bool
	// ExpectWalkLog: when true, Assert expects home walk artifacts after Run.
	ExpectWalkLog bool

	// --- P4 walk consume ---

	// Consume: Run performs a cold Scan, optional FS mutations, then a second
	// Scan (Refresh=false, WarmRefreshBudget=-1) that must drive walk-log
	// consume under the selected sync budget.
	Consume bool

	// LastScanEnd is injected as Options.LastScanEnd on the second Scan when
	// SetLastScanEnd is true. Product field required for P4.
	LastScanEnd    time.Time
	SetLastScanEnd bool

	// NowAt is returned by Options.Now on the second Scan when SetNow is true.
	NowAt  time.Time
	SetNow bool

	// DeleteRelPaths: after cold, os.RemoveAll under Roots[0] for each rel path.
	DeleteRelPaths []string
	// AddRepoRelPaths: after cold, plant fake .git mains at these rel paths.
	AddRepoRelPaths []string

	// BudgetOnly: skip Scan; only resolve WalkConsumeSyncBudget(DeltaAge).
	// Used by pure budget-tier leaves (10s–60s and structural asserts).
	BudgetOnly bool
	// DeltaAge is the synthetic (Now − LastScanEnd) for BudgetOnly or for
	// documenting the tier when also running Consume.
	DeltaAge time.Duration
}

type Response struct {
	Repos      []scan_repo.Repo
	RootErrors []scan_repo.RootError
	// WalkLogPath / CursorPath are the expected on-disk locations under home/.
	WalkLogPath string
	CursorPath  string
	// WalkLogOK is true when walk.jsonl exists and was readable after Run.
	WalkLogOK bool
	// WalkEvents are parsed non-empty lines of walk.jsonl (order preserved).
	WalkEvents []WalkEvent
	// WalkLogSize is the byte length of walk.jsonl after Run (0 if missing).
	WalkLogSize int64
	// CursorOK is true when walk.cursor.json exists and was parseable.
	CursorOK bool
	// Cursor is the parsed cursor document (Offset meaningful when CursorOK).
	Cursor WalkCursor

	// After cold (Consume path): snapshots before second Scan.
	ColdWalkLogSize   int64
	ColdCursorOffset  int64
	ColdWalkLogOK     bool
	ColdCursorOK      bool
	ColdEventCount    int
	// Second Scan result repos (same as Repos when Consume); kept explicit.
	ConsumeRepos []scan_repo.Repo

	// SelectedBudget is WalkConsumeSyncBudget(DeltaAge) when BudgetOnly or when
	// SetLastScanEnd+SetNow allow computing delta = NowAt − LastScanEnd.
	SelectedBudget time.Duration
	// SelectedBudgetOK is true when the product budget helper was callable.
	SelectedBudgetOK bool

	DebugOut string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	walkLogPath := filepath.Join(req.CacheRoot, walkUniverseHome, "walk.jsonl")
	cursorPath := filepath.Join(req.CacheRoot, walkUniverseHome, "walk.cursor.json")

	resp := &Response{
		WalkLogPath: walkLogPath,
		CursorPath:  cursorPath,
	}

	// Pure budget selection (no Scan). Product must export:
	//   WalkConsumeSyncBudget(sinceLast time.Duration) time.Duration
	// RED until implemented.
	if req.BudgetOnly {
		budget, ok := callWalkConsumeSyncBudget(req.DeltaAge)
		resp.SelectedBudget = budget
		resp.SelectedBudgetOK = ok
		if !ok {
			return resp, fmt.Errorf("WalkConsumeSyncBudget not available: product must export WalkConsumeSyncBudget(time.Duration) time.Duration for P4 adaptive walk-consume budget")
		}
		return resp, nil
	}

	var stderr bytes.Buffer

	// --- Scan 1: cold (or single-shot for non-Consume leaves) ---
	opts1 := scan_repo.Options{
		Roots:     req.Roots,
		CacheRoot: req.CacheRoot,
		NoCache:   req.NoCache,
		Refresh:   req.Refresh,
		Debug:     req.Debug,
	}
	if req.Debug {
		opts1.Stderr = &stderr
	}
	result1, err := scan_repo.Scan(context.Background(), opts1)
	if err != nil {
		return nil, err
	}
	resp.Repos = result1.Repos
	resp.RootErrors = result1.RootErrors
	resp.DebugOut = stderr.String()

	if !req.Consume {
		if fillErr := fillWalkState(resp, walkLogPath, cursorPath); fillErr != nil {
			return nil, fillErr
		}
		return resp, nil
	}

	// Snapshot cold seal state before mutations / second Scan.
	if fillErr := fillWalkState(resp, walkLogPath, cursorPath); fillErr != nil {
		return nil, fillErr
	}
	resp.ColdWalkLogSize = resp.WalkLogSize
	resp.ColdCursorOffset = resp.Cursor.Offset
	resp.ColdWalkLogOK = resp.WalkLogOK
	resp.ColdCursorOK = resp.CursorOK
	resp.ColdEventCount = len(resp.WalkEvents)

	// Post-cold filesystem mutations (gone / new repo fixtures).
	if len(req.Roots) == 0 {
		return nil, fmt.Errorf("Consume requires Roots")
	}
	root := req.Roots[0]
	for _, rel := range req.DeleteRelPaths {
		p := filepath.Join(root, rel)
		if rmErr := os.RemoveAll(p); rmErr != nil {
			return nil, fmt.Errorf("DeleteRelPaths %q: %w", rel, rmErr)
		}
	}
	for _, rel := range req.AddRepoRelPaths {
		p := filepath.Join(root, rel)
		if mkErr := os.MkdirAll(p, 0755); mkErr != nil {
			return nil, mkErr
		}
		gitDir := filepath.Join(p, ".git")
		if mkErr := os.MkdirAll(filepath.Join(gitDir, "objects"), 0755); mkErr != nil {
			return nil, mkErr
		}
	}

	// Selected budget from injected times (helps Assert without re-deriving).
	if req.SetLastScanEnd && req.SetNow {
		delta := req.NowAt.Sub(req.LastScanEnd)
		if budget, ok := callWalkConsumeSyncBudget(delta); ok {
			resp.SelectedBudget = budget
			resp.SelectedBudgetOK = true
		}
	}

	// --- Scan 2: warm + walk consume ---
	// WarmRefreshBudget < 0 disables unit warm-refresh so discoveries/gone
	// observed here must come from walk.jsonl consume (P4), not unit rewalk.
	var stderr2 bytes.Buffer
	opts2 := scan_repo.Options{
		Roots:             req.Roots,
		CacheRoot:         req.CacheRoot,
		NoCache:           false,
		Refresh:           false,
		Debug:             req.Debug,
		WarmRefreshBudget: -1,
	}
	if req.Debug {
		opts2.Stderr = &stderr2
	}
	if req.SetLastScanEnd {
		// Product Options must include LastScanEnd time.Time (P4).
		opts2.LastScanEnd = req.LastScanEnd
	}
	if req.SetNow {
		nowAt := req.NowAt
		opts2.Now = func() time.Time { return nowAt }
	}
	result2, err2 := scan_repo.Scan(context.Background(), opts2)
	if err2 != nil {
		return nil, err2
	}
	resp.Repos = result2.Repos
	resp.ConsumeRepos = result2.Repos
	resp.RootErrors = result2.RootErrors
	resp.DebugOut += stderr2.String()

	if fillErr := fillWalkState(resp, walkLogPath, cursorPath); fillErr != nil {
		return nil, fillErr
	}
	return resp, nil
}

// callWalkConsumeSyncBudget invokes the product budget helper.
// Prefer exported WalkConsumeSyncBudget; fall back to TestExported_ if present.
func callWalkConsumeSyncBudget(sinceLast time.Duration) (time.Duration, bool) {
	return scan_repo.WalkConsumeSyncBudget(sinceLast), true
}

func fillWalkState(resp *Response, walkLogPath, cursorPath string) error {
	resp.WalkLogOK = false
	resp.WalkLogSize = 0
	resp.WalkEvents = nil
	resp.CursorOK = false
	resp.Cursor = WalkCursor{}

	if st, statErr := os.Stat(walkLogPath); statErr == nil && !st.IsDir() {
		resp.WalkLogOK = true
		resp.WalkLogSize = st.Size()
		events, parseErr := parseWalkJSONL(walkLogPath)
		if parseErr != nil {
			return fmt.Errorf("parse walk.jsonl: %w", parseErr)
		}
		resp.WalkEvents = events
	}
	if raw, readErr := os.ReadFile(cursorPath); readErr == nil {
		var cur WalkCursor
		if jsonErr := json.Unmarshal(raw, &cur); jsonErr != nil {
			return fmt.Errorf("parse walk.cursor.json: %w", jsonErr)
		}
		resp.CursorOK = true
		resp.Cursor = cur
	}
	return nil
}

func parseWalkJSONL(path string) ([]WalkEvent, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var events []WalkEvent
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var ev WalkEvent
		if err := json.Unmarshal(line, &ev); err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNo, err)
		}
		events = append(events, ev)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return events, nil
}
```
