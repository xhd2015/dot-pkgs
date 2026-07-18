# Scenario

**Feature**: cold Scan writes visit events for walked dirs, seals gen_end 1, sets cursor

```
# workspace: non-repo root with intermediate dir + one main repo
workspace/
  notes/           (plain dir)
  projects/
    alpha/         (fake .git main)
  -> Scan(CacheRoot, NoCache=false)  # first cold; no prior walk log
  -> Result discovers alpha
  -> home/walk.jsonl:
       visit for abs(workspace), abs(notes), abs(projects), abs(alpha) (at least)
       last event: {"op":"gen_end","gen":1}
  -> home/walk.cursor.json: {"offset": <len(walk.jsonl)>}
```

## Steps

1. Create workspace with `notes/` (non-repo), `projects/alpha/` (main repo).
2. Set `req.Roots` to the workspace; keep `NoCache=false`.
3. Do **not** pre-seed walk log or mirror — Run is the cold Scan under test.

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
	req.NoCache = false
	req.Refresh = false
	req.ExpectWalkLog = true
	return nil
}
```
