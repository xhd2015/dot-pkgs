## Steps
- Write a history file with root + worktree entry.
- Write an alias pointing to the root.
- Run --picker-dump to verify alias annotation appears on the root entry.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	root := filepath.Join(req.WorkRoot, "repo")
	wt := filepath.Join(req.WorkRoot, "feature")
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
	writeAliasesFile(t, req.ConfigHome, map[string]string{
		"myproj": root,
	})

	req.Args = []string{"--picker-dump"}
	return nil
}
```
