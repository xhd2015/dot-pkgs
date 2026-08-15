# Scenario

External main that is also latest — not duplicated.

external main = latest → not duplicated

## Steps
- Write history with chain: root → worktree → plain-move (no second worktree).
- The plain-move destination (dst) is both external main AND latest.
- Create dst only; root and wt1 stay dead.
- Run --picker-list to verify dst shows as (external main) and is not duplicated.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := filepath.Join(req.WorkRoot, "repo")
	wt1 := filepath.Join(req.WorkRoot, "feature-a")
	dst := filepath.Join(req.WorkRoot, "repo-moved")

	mkdirAll(t, dst)

	hf := HistoryFile{
		Version: "1.1",
		Projects: map[string]ProjectEntry{
			root: {
				Locations: []LocationEntry{
					{Path: root},
					{Path: wt1, Git: &GitInfo{Type: "worktree", MainRepo: root, Branch: "feature-a"}},
					{Path: dst},
				},
			},
		},
	}
	writeHistoryFile(t, req.ConfigHome, hf)

	req.Args = []string{"--picker-list"}
	return nil
}
```
