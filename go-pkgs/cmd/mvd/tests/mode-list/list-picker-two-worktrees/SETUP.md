# Scenario

Picker dump shows root + 2 worktrees (3 entries).

mvd -w repo wt1; mvd -w repo wt2 → [(repo), (wt1 w:wt1), (wt2 w:wt2)]
mvd --picker-list → shows 3 entries

## Steps
- Write a history file with one entry: root path + two worktree locations with git metadata.
- Run --picker-list to verify all three paths appear.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := filepath.Join(req.WorkRoot, "repo")
	wt1 := filepath.Join(req.WorkRoot, "feature-a")
	wt2 := filepath.Join(req.WorkRoot, "feature-b")
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
