# Scenario

No marker for plain alive entry without worktree or alias.

plain alive entry → (no marker)

## Steps
- Write history with a single root location, no worktree, no alias.
- Create the directory so it is alive.
- Run --picker-list to verify no marker is shown.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	root := filepath.Join(req.WorkRoot, "repo")
	mkdirAll(t, root)

	hf := HistoryFile{
		Version: "1.1",
		Projects: map[string]ProjectEntry{
			root: {
				Locations: []LocationEntry{
					{Path: root},
				},
			},
		},
	}
	writeHistoryFile(t, req.ConfigHome, hf)

	req.Args = []string{"--picker-list"}
	return nil
}
```
