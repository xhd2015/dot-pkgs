## Steps
- Write a history file with one entry: root path + two worktree locations with git metadata.
- Run --picker-dump to verify all three paths appear.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
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

	req.Args = []string{"--picker-dump"}
	return nil
}
```
