# Scenario

**Feature**: pure `ChildrenIndex` and `Descendants` over an in-memory process table

```
[]Proc fixture
  -> ChildrenIndex -> map[ppid][]child (sorted)
  -> Descendants(root, maxDepth) -> BFS []Proc including root
```

## Preconditions

- Leaves supply `req.FixtureProcs` and either children-index or descendants
  fields (`RootPID`, `MaxDepth`).
- No live process listing.

## Steps

1. Grouping does not set `Op` (leaves set `children-index` or `descendants`).
2. Provide a shared linear-chain fixture helper for depth leaves.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

// linearChainFixture: 1→2→3→4→5 (PPID chain under 0).
func linearChainFixture() []FixtureProc {
	return []FixtureProc{
		{PID: 1, PPID: 0, Cmd: "root"},
		{PID: 2, PPID: 1, Cmd: "child"},
		{PID: 3, PPID: 2, Cmd: "grand"},
		{PID: 4, PPID: 3, Cmd: "great"},
		{PID: 5, PPID: 4, Cmd: "deep"},
	}
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	// Tree leaves set Op + fixture; grouping only documents shared helpers.
	if req.FixtureProcs == nil {
		req.FixtureProcs = []FixtureProc{}
	}
	return nil
}
```
