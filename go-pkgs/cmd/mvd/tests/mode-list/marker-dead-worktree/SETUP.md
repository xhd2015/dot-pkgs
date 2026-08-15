# Scenario

Dead worktree marker for missing worktree path.

create then delete worktree → (dead worktree)

## Steps
- Write history with root + 1 worktree, but only create the root directory (worktree dir is dead).
- Run --picker-list to verify `(main)` on root and `(dead worktree)` on worktree.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := filepath.Join(req.WorkRoot, "repo")
	wt := filepath.Join(req.WorkRoot, "feature")
	mkdirAll(t, root)

	hf := HistoryFile{
		Version: "1.1",
		Projects: map[string]ProjectEntry{
			root: {
				Locations: []LocationEntry{
					{Path: root},
					{Path: wt, Git: &GitInfo{Type: "worktree", MainRepo: root, Branch: "feature"}},
				},
			},
		},
	}
	writeHistoryFile(t, req.ConfigHome, hf)

	req.Args = []string{"--picker-list"}
	return nil
}
```
