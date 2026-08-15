# Scenario

External main path shown (external main). Bug fix for root→WT→plain→WT.

complex chain with external main → shows (external main)

## Steps
- Write history with chain: root → worktree → plain-move → worktree.
- This tests the bug fix: the plain-move destination (dst) must be shown as (external main).
- Create dst and wt2 so they are alive; root and wt1 stay dead (not created on FS).
- Run --picker-list to verify all 4 entries show with correct markers.

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

	mkdirAll(t, dst)
	mkdirAll(t, wt2)

	hf := HistoryFile{
		Version: "1.1",
		Projects: map[string]ProjectEntry{
			root: {
				Locations: []LocationEntry{
					{Path: root},
					{Path: wt1, Git: &GitInfo{Type: "worktree", MainRepo: root, Branch: "feature-a"}},
					{Path: dst},
					{Path: wt2, Git: &GitInfo{Type: "worktree", MainRepo: dst, Branch: "feature-b"}},
				},
			},
		},
	}
	writeHistoryFile(t, req.ConfigHome, hf)

	req.Args = []string{"--picker-list"}
	return nil
}
```
