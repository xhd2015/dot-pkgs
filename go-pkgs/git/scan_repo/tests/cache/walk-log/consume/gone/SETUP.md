# Scenario

**Feature**: re-list of a removed visit path appends a gone event

```
# cold visits notes/ then notes removed
workspace/
  notes/            (plain dir — visited on cold)
  projects/alpha/   (main)
  -> cold Scan: visit abs(notes) among others; gen_end 1
  -> delete notes/
  -> second Scan(delta>=60s, WarmRefreshBudget=-1)
  -> after gen_end 1: {"op":"gone","path":abs(notes)} (at least)
  -> gen_end 2 still sealed when generation consumed
```

## Steps

1. Plant notes/ + projects/alpha.
2. Set `DeleteRelPaths=["notes"]` so Run removes notes after cold.
3. Default full-budget clocks.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	mkdirAll(t, filepath.Join(root, "notes"))
	alpha := filepath.Join(root, "projects", "alpha")
	mkdirAll(t, alpha)
	fakeGitRepo(t, alpha)
	req.Roots = []string{root}
	req.Consume = true
	req.DeleteRelPaths = []string{"notes"}
	return nil
}
```
