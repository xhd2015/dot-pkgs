# Scenario

Alias marker on entry that was moved (alias follows moves).

mvd --add repo; mvd --add-alias repo al → [(repo)]
mvd repo dst → [(repo), (dst/repo)]
markers → (aliases: al)

## Steps
- Write a history file with one entry: a plain move chain (root → moved, no git metadata) with an alias "dp".
- The directory for the latest path must exist so it appears as alive.
- Run --picker-list to verify `(aliases: dp)` marker appears on the latest (moved) entry.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := filepath.Join(req.WorkRoot, "repo")
	moved := filepath.Join(req.WorkRoot, "repo-moved")
	mkdirAll(t, moved)

	hf := HistoryFile{
		Version: "1.1",
		Projects: map[string]ProjectEntry{
			root: {
				Locations: []LocationEntry{
					{Path: root},
					{Path: moved},
				},
				Aliases: []string{"dp"},
			},
		},
	}
	writeHistoryFile(t, req.ConfigHome, hf)

	req.Args = []string{"--picker-list"}
	return nil
}
```
