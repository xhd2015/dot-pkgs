# Scenario

Root (main) + 2 worktrees markers, alive.

mvd -w repo wt1; mvd -w repo wt2 → [(repo), (wt1 w:wt1), (wt2 w:wt2)]
markers → (main), (worktree), (worktree)

## Steps
- Write history with root + 2 worktree locations with git metadata.
- Create all three directories so they are alive.
- Run --picker-list to verify (main) and two (worktree) markers.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := filepath.Join(req.WorkRoot, "repo")
	wt1 := filepath.Join(req.WorkRoot, "feature-a")
	wt2 := filepath.Join(req.WorkRoot, "feature-b")
	mkdirAll(t, root)
	mkdirAll(t, wt1)
	mkdirAll(t, wt2)

	hf := HistoryFile{
		Version: "1.1",
		Projects: map[string]ProjectEntry{
			root: {
				Locations: []LocationEntry{
					{Path: root},
					{Path: wt1, Git: &GitInfo{Type: "worktree", MainRepo: root, Branch: "feature-a"}},
					{Path: wt2, Git: &GitInfo{Type: "worktree", MainRepo: root, Branch: "feature-b"}},
				},
			},
		},
	}
	writeHistoryFile(t, req.ConfigHome, hf)

	req.Args = []string{"--picker-list"}
	return nil
}
```
