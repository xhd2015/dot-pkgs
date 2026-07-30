# Scenario

**Feature**: shared `proc` primitives for list, open-files, descendants, Alive

```
# pure parsers (fixture bytes)
PS/lsof sample bytes -> ParsePSOutput / ParseLsofFn -> rows / paths

# tree helpers (in-memory table)
[]Proc -> ChildrenIndex | Descendants(root, maxDepth) -> index / BFS set

# injectable live surfaces
Options hooks -> List / OpenFiles / Alive -> fixture results (no real ps/lsof)
```

## Preconditions

- Package under test: `github.com/xhd2015/dot-pkgs/go-pkgs/proc` (absent until
  implementer — Classic TDD RED).
- All leaves are L2 in-process: pure parsers or Options inject; no live
  agent-pro, kool, grok, or codex session paths.
- Parallel-safe: inject via `proc.Options` fields only (no package globals).

## Steps

1. Root `Setup` clears request fields to zero defaults for the leaf chain.
2. Grouping/leaf `Setup` sets `req.Op` and capability-specific inputs.
3. Root `Run` dispatches on `req.Op` to the locked API surface.
4. Leaf `Assert` checks response fields only (no re-implementation of parsers).

## Context

- Locked API (implementer contract): `Proc`, `Options`, `List`, `ParsePSOutput`,
  `OpenFiles`, `ParseLsofFn`, `ChildrenIndex`, `Descendants`, `Alive`.
- `Descendants`: include root when present; `maxDepth <= 0` → 16; missing root
  → empty slice; BFS with children sorted by PID.
- `Alive`: `pid <= 0` → false even if inject is set.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	// Zero defaults so leaves only set what they need; Op is required by Run.
	req.Op = ""
	req.PSOutput = nil
	req.LsofOutput = nil
	req.FixtureProcs = nil
	req.RootPID = 0
	req.MaxDepth = 0
	req.ListInject = nil
	req.OpenFilesPID = 0
	req.OpenFilesInject = nil
	req.AlivePID = 0
	req.AliveInject = false
	req.AliveUseInject = false
	return nil
}

func assertProcEqual(t *testing.T, got, want FixtureProc) {
	t.Helper()
	if got.PID != want.PID || got.PPID != want.PPID || got.Cmd != want.Cmd {
		t.Fatalf("proc mismatch: got={PID:%d PPID:%d Cmd:%q} want={PID:%d PPID:%d Cmd:%q}",
			got.PID, got.PPID, got.Cmd, want.PID, want.PPID, want.Cmd)
	}
}

func assertProcsEqual(t *testing.T, got, want []FixtureProc) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("procs len=%d want %d\n got=%v\nwant=%v", len(got), len(want), got, want)
	}
	for i := range want {
		assertProcEqual(t, got[i], want[i])
	}
}

func assertStringsEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("strings len=%d want %d\n got=%v\nwant=%v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("strings[%d]=%q want %q (full got=%v want=%v)", i, got[i], want[i], got, want)
		}
	}
}

func assertIntSliceEqual(t *testing.T, got, want []int) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("ints len=%d want %d\n got=%v\nwant=%v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ints[%d]=%d want %d (full got=%v want=%v)", i, got[i], want[i], got, want)
		}
	}
}
```
