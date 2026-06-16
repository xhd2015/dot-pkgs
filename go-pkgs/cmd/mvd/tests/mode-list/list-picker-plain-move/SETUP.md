# Scenario

Picker dump for plain move shows only the latest (1 entry).

mvd src dst → [(src), (dst/src)]
mvd --picker-list → shows 1 entry

## Steps
- Write a history file with one entry: a plain move chain (root → moved, no git metadata).
- Run --picker-list to verify only the latest location appears (regression: plain moves should not show the old root).

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	root := filepath.Join(req.WorkRoot, "repo")
	moved := filepath.Join(req.WorkRoot, "repo-moved")
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
