# proc — process list, open-files, descendants, Alive

Classic TDD doctests for plan phase **P1**: package
`github.com/xhd2015/dot-pkgs/go-pkgs/proc` — **RED** until implementer lands
production sources under `go-pkgs/proc/`.

Shared process primitives: parse/list processes, parse/list open files, build
children index and BFS descendants, and probe Alive — all testable with pure
parsers and injectable Options (no live agent, no kool/grok/codex session paths).

**Out of scope:** `procresolve`, agent-pro CLI, session classification, live
`ps`/`lsof` e2e (optional; inject fixtures cover P1 exit).

## Version

0.0.2

## DSN (Domain Specific Notion)

Shared process primitives package for callers that need process rows, open-file
paths, descendant trees, and liveness without shelling into agent tooling.

### Participants

- **Caller** — library client that needs process table rows, open files for a
  pid, a descendant subtree, or a liveness check.
- **`List` / `ParsePSOutput`** — `List(opts Options) []Proc` returns a process
  snapshot. When `opts.List != nil`, use the inject; otherwise live
  `ps -ax -o pid=,ppid=,command=` (fallback `ps -x …`). On live failure → empty
  (not panic). `ParsePSOutput(out []byte) []Proc` is the pure parser for fixture
  tests.
- **`OpenFiles` / `ParseLsofFn`** — `OpenFiles(pid, opts) []string` returns open
  paths. When `opts.OpenFiles != nil`, use inject; otherwise live
  `lsof -p <pid> -Fn`. On failure → empty. `ParseLsofFn` is pure: lines starting
  with `n` → path; skip empty and `"/"`; keep unique absolute paths (first-seen
  order).
- **`ChildrenIndex` / `Descendants`** — pure tree helpers over `[]Proc`.
  `ChildrenIndex` maps PPID → child PIDs (each slice sorted ascending).
  `Descendants(rootPID, procs, maxDepth)` BFS from root, **includes root** when
  present; `maxDepth <= 0` means default **16**; stable order = BFS, children
  expanded by ascending PID; missing root → empty slice (no error).
- **`Alive`** — `Alive(pid, opts) bool`. `pid <= 0` → always `false` (before
  inject/live). When `opts.Alive != nil`, use inject; otherwise Unix signal-0 /
  `FindProcess` probe.
- **`Options`** — injectable hooks only (parallel-safe; no package globals):

```text
Options
  List      func() []Proc
  OpenFiles func(pid int) []string
  Alive     func(pid int) bool

Proc
  PID, PPID int
  Cmd string   // full command line when available
```

### Behaviors

- **Parse PS** — leading pid/ppid integers; remainder is Cmd; skip empty lines
  and lines that lack two leading ints; never panic.
- **Parse lsof `-Fn`** — only `n…` name fields that are absolute paths other
  than `/`; unique; junk / non-`n` lines ignored.
- **Descendants BFS** — depth(root)=0; expand while depth < maxDepth; include
  every visited node in visit order.
- **List / OpenFiles inject** — when hook set, return inject result as-is (no
  re-parse).
- **Alive inject** — when hook set and pid > 0, return inject result; pid ≤ 0
  still false.

## Decision Tree

```
proc/tests/core/
├── parse-ps/                     # pure ParsePSOutput
│   ├── multi-line/               # sample ps table → PID/PPID/Cmd rows
│   └── skip-invalid/             # empty + garbage lines skipped, no panic
├── parse-lsof/                   # pure ParseLsofFn
│   ├── n-paths/                  # n-paths → unique absolute paths
│   └── skip-junk/                # n/, empty n, non-n, relative skipped
├── tree/                         # ChildrenIndex + Descendants
│   ├── children-index/           # PPID → sorted child PIDs
│   ├── linear-chain/             # 1→2→3 depth-limited BFS set
│   ├── max-depth-default/        # maxDepth<=0 → default 16, finds deep chain
│   └── missing-root/             # root absent → empty
├── list/                         # List with Options.List inject
│   └── inject-fixture/           # custom List returns fixture rows
├── open-files/                   # OpenFiles with Options.OpenFiles inject
│   └── inject-paths/             # custom returns paths as-is
└── alive/                        # Alive
    ├── pid-nonpositive/          # pid<=0 → false
    └── inject-true-false/        # Options.Alive respected for pid>0
```

Parameter ranking (most → least significant):

1. **Capability** — parse-ps / parse-lsof / tree / list / open-files / alive
2. **Input shape** — valid fixture vs junk/edge vs inject vs depth/root presence

## Test Index

| # | Leaf | Op | Description |
|---|------|-----|-------------|
| 1 | `parse-ps/multi-line` | parse-ps | Multi-line `ps` sample → correct PID/PPID/Cmd rows |
| 2 | `parse-ps/skip-invalid` | parse-ps | Empty + garbage lines skipped; no panic; only valid rows |
| 3 | `parse-lsof/n-paths` | parse-lsof | `-Fn` sample → unique absolute paths (first-seen order) |
| 4 | `parse-lsof/skip-junk` | parse-lsof | `n/`, empty `n`, non-`n`, relative → omitted |
| 5 | `tree/children-index` | children-index | PPID → child PIDs sorted ascending |
| 6 | `tree/linear-chain` | descendants | Chain 1→2→3 with maxDepth=1 → PIDs [1,2] only |
| 7 | `tree/max-depth-default` | descendants | maxDepth=0 → default 16; deep chain all found |
| 8 | `tree/missing-root` | descendants | Root PID absent from table → empty slice |
| 9 | `list/inject-fixture` | list | `Options.List` returns fixture; List uses inject |
| 10 | `open-files/inject-paths` | open-files | `Options.OpenFiles` returns paths used as-is |
| 11 | `alive/pid-nonpositive` | alive | pid ≤ 0 → false |
| 12 | `alive/inject-true-false` | alive | `Options.Alive` true/false for positive pids |

## How to Run

```sh
# from agent-pro worktree root, or cd into go-pkgs
doctest vet ./external/dot-pkgs-master-2026-07-30-1/go-pkgs/proc/tests/core
doctest test ./external/dot-pkgs-master-2026-07-30-1/go-pkgs/proc/tests/core

doctest test -v ./external/dot-pkgs-master-2026-07-30-1/go-pkgs/proc/tests/core/parse-ps/multi-line
doctest test -v ./external/dot-pkgs-master-2026-07-30-1/go-pkgs/proc/tests/core/tree/linear-chain
doctest test -v ./external/dot-pkgs-master-2026-07-30-1/go-pkgs/proc/tests/core/alive/inject-true-false
```

Classic TDD: expect **RED** (compile or assert failure) until
`go-pkgs/proc` production package exists and implements the locked API.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/proc"
)

// FixtureProc mirrors proc.Proc for request fixtures without forcing every
// leaf to import field order trivia.
type FixtureProc struct {
	PID  int
	PPID int
	Cmd  string
}

// Request is filled root→leaf. Op selects the capability under test.
type Request struct {
	// Op: parse-ps | parse-lsof | children-index | descendants | list | open-files | alive
	Op string

	// Pure parser inputs
	PSOutput   []byte
	LsofOutput []byte

	// Process table for tree ops
	FixtureProcs []FixtureProc
	RootPID      int
	MaxDepth     int

	// List inject (Op=list)
	ListInject []FixtureProc

	// OpenFiles inject (Op=open-files)
	OpenFilesPID    int
	OpenFilesInject []string

	// Alive (Op=alive)
	AlivePID       int
	AliveInject    bool
	AliveUseInject bool
}

// Response observes the selected op result.
type Response struct {
	Procs    []FixtureProc
	Paths    []string
	Children map[int][]int
	Alive    bool
}

func toProc(fps []FixtureProc) []proc.Proc {
	out := make([]proc.Proc, 0, len(fps))
	for _, fp := range fps {
		out = append(out, proc.Proc{PID: fp.PID, PPID: fp.PPID, Cmd: fp.Cmd})
	}
	return out
}

func fromProc(ps []proc.Proc) []FixtureProc {
	out := make([]FixtureProc, 0, len(ps))
	for _, p := range ps {
		out = append(out, FixtureProc{PID: p.PID, PPID: p.PPID, Cmd: p.Cmd})
	}
	return out
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	resp := &Response{}

	switch req.Op {
	case "parse-ps":
		resp.Procs = fromProc(proc.ParsePSOutput(req.PSOutput))
	case "parse-lsof":
		resp.Paths = proc.ParseLsofFn(req.LsofOutput)
	case "children-index":
		resp.Children = proc.ChildrenIndex(toProc(req.FixtureProcs))
	case "descendants":
		resp.Procs = fromProc(proc.Descendants(req.RootPID, toProc(req.FixtureProcs), req.MaxDepth))
	case "list":
		opts := proc.Options{
			List: func() []proc.Proc {
				return toProc(req.ListInject)
			},
		}
		resp.Procs = fromProc(proc.List(opts))
	case "open-files":
		paths := append([]string(nil), req.OpenFilesInject...)
		opts := proc.Options{
			OpenFiles: func(pid int) []string {
				if pid == req.OpenFilesPID {
					return paths
				}
				return nil
			},
		}
		resp.Paths = proc.OpenFiles(req.OpenFilesPID, opts)
	case "alive":
		opts := proc.Options{}
		if req.AliveUseInject {
			want := req.AliveInject
			opts.Alive = func(pid int) bool {
				return want
			}
		}
		resp.Alive = proc.Alive(req.AlivePID, opts)
	default:
		t.Fatalf("unknown Op %q", req.Op)
	}
	return resp, nil
}
```
