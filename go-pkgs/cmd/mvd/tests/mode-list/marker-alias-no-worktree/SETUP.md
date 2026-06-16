# Scenario

Alias marker on plain entry without worktree.

mvd --add repo; mvd --add-alias repo al → [(repo)]
markers → (aliases: al)

## Steps
- Write history with a single root location (no worktree) and an alias "myproj".
- Create the directory so it is alive.
- Run --picker-list to verify `(aliases: myproj)` marker shows.

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
				Aliases: []string{"myproj"},
			},
		},
	}
	writeHistoryFile(t, req.ConfigHome, hf)

	req.Args = []string{"--picker-list"}
	return nil
}
```
