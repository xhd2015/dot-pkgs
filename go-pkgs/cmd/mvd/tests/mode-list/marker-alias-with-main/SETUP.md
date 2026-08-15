# Scenario

Combined marker (main, aliases: ...) when root has alias + is main.

mvd --add repo; mvd --add-alias repo al → [(repo)]
markers → (main, aliases: al)

## Steps
- Write history with root + worktree + alias ("myproj").
- Create both directories so they are alive.
- Run --picker-list to verify root gets combined marker `(main, aliases: myproj)`.
- The worktree line must NOT contain the alias text.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := filepath.Join(req.WorkRoot, "repo")
	wt := filepath.Join(req.WorkRoot, "feature")
	mkdirAll(t, root)
	mkdirAll(t, wt)

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
