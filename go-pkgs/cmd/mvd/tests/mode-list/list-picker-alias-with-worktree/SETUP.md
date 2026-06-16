# Scenario

Alias annotation on root entry, not worktree.

mvd --add repo; mvd --add-alias repo al → [(repo)]
mvd -w repo wt → [(repo), (wt w:wt)]
mvd --picker-list → alias on root, not worktree

## Steps
- Write a history file with root + worktree entry.
- Write an alias pointing to the root.
- Run --picker-list to verify alias annotation appears on the root entry.

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
				Aliases: []string{"myproj"},
			},
		},
	}
	writeHistoryFile(t, req.ConfigHome, hf)

	req.Args = []string{"--picker-list"}
	return nil
}
```
