# Scenario

Remove a dead worktree path from a chain via chain-path resolution.

Sibling case from the same bug report: dead worktree in a long chain (agent-pro-fix-trace topology).

Chain: root → wt1 → dst → wt2 where wt2 is dead (dst remains alive).

## Steps
- Seed history with the 4-location chain.
- Create root, wt1, dst directories but NOT wt2 (wt2 is dead on disk).
- Run `mvd --rm <wt2>` using the absolute path.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	root := filepath.Join(req.WorkRoot, "repo")
	wt1 := filepath.Join(req.WorkRoot, "feature-a")
	dst := filepath.Join(req.WorkRoot, "repo-moved")
	wt2 := filepath.Join(req.WorkRoot, "feature-b")

	mkdirAll(t, root)
	mkdirAll(t, wt1)
	mkdirAll(t, dst)

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

	req.Args = []string{"--rm", wt2}
	return nil
}
```