# Scenario

Dead marker for dead plain entry (no worktree).

dead plain entry → (dead)

## Steps
- Write history with a plain move chain (root → moved, no git metadata).
- Create only the root directory; moved does not exist.
- Run --picker-list to verify moved shows `(dead)` marker (only latest shown for plain entries).

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := filepath.Join(req.WorkRoot, "repo")
	moved := filepath.Join(req.WorkRoot, "repo-moved")
	mkdirAll(t, root)

	hf := HistoryFile{
		Version: "1.1",
		Projects: map[string]ProjectEntry{
			root: {
				Locations: []LocationEntry{
					{Path: root},
					{Path: moved},
				},
			},
		},
	}
	writeHistoryFile(t, req.ConfigHome, hf)

	req.Args = []string{"--picker-list"}
	return nil
}
```
