# Scenario

v3.0 moves round-trip for external-main + worktree topology (agent-pro chain).

Chain: root → wt1 (worktree) → dst (external main) → wt2 (worktree from dst)

## Steps
- Seed v3.0 history with explicit `moves` (no `locations`).
- Create all directories including dst (dst is alive).
- Run `--picker-list` to verify markers and load round-trip.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := filepath.Join(req.WorkRoot, "repo")
	wt1 := filepath.Join(req.WorkRoot, "feature-a")
	dst := filepath.Join(req.WorkRoot, "repo-moved")
	wt2 := filepath.Join(req.WorkRoot, "feature-b")

	mkdirAll(t, root)
	mkdirAll(t, wt1)
	mkdirAll(t, dst)
	mkdirAll(t, wt2)

	hf := HistoryFile{
		Version: "3.0",
		Projects: map[string]ProjectEntry{
			root: {
				Root: root,
				Moves: []MoveEntry{
					{From: root, FromType: "main", To: wt1, ToType: "worktree", Branch: "feature-a"},
					{From: root, FromType: "main", To: dst, ToType: "main"},
					{From: dst, FromType: "main", To: wt2, ToType: "worktree", Branch: "feature-b"},
				},
			},
		},
	}
	writeHistoryFile(t, req.ConfigHome, hf)

	req.Args = []string{"--picker-list"}
	return nil
}
```