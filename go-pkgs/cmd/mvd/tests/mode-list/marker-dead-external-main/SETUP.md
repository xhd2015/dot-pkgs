# Scenario

Dead external main marker.

dead external main path → (dead external main)

## Steps
- Write history with chain: root → wt1 → dst → wt2.
- Create root, wt1, wt2 directories but NOT dst (dst is dead).
- Run --picker-list to verify dst shows as `(dead external main)`.

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
