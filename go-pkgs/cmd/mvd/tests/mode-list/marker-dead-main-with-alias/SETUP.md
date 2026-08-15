# Scenario

Combined dead marker for dead root with alias.

dead main with alias → (dead main, aliases: ...)

## Steps
- Write history with root (with alias "myproj") + 1 worktree, but only create the worktree.
- Root is dead AND main AND has alias → expects `(dead main, aliases: myproj)`.
- Worktree is alive → expects `(worktree)`.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := filepath.Join(req.WorkRoot, "repo")
	wt := filepath.Join(req.WorkRoot, "feature")
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
