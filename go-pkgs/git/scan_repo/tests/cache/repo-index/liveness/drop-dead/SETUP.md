# Scenario

**Feature**: ApplyLiveness removes missing-.git paths and keeps live fake repos

```
# seed index
live-repo/.git present + dead-repo/ with no .git (path may exist or not)
  -> ApplyLiveness
  -> live-repo entry remains; dead-repo entry gone
```

## Steps

1. Create workspace dirs: `live-repo/` with `fakeGitRepo`, and `dead-repo/` as
   a plain directory (no `.git`).
2. Build `req.Index` with both paths as `main` entries under universe `home`.
3. Do not write `repos.json` — liveness is pure in-memory filter for this leaf.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := t.TempDir()
	live := filepath.Join(root, "live-repo")
	dead := filepath.Join(root, "dead-repo")
	mkdirAll(t, live)
	fakeGitRepo(t, live)
	mkdirAll(t, dead) // exists but no .git

	liveAbs := absPath(t, live)
	deadAbs := absPath(t, dead)

	req.Universe = "home"
	req.Index = scan_repo.RepoIndex{
		Version:   1,
		Universe:  "home",
		Base:      absPath(t, root),
		UpdatedAt: "2026-07-15T14:00:00Z",
		Repos: []scan_repo.RepoIndexEntry{
			{
				Path:     deadAbs,
				RepoType: "main",
				GitDir:   filepath.Join(deadAbs, ".git"),
				Depth:    1,
				SeenAt:   "2026-07-15T13:00:00Z",
			},
			{
				Path:     liveAbs,
				RepoType: "main",
				GitDir:   filepath.Join(liveAbs, ".git"),
				Depth:    1,
				SeenAt:   "2026-07-15T13:01:00Z",
			},
		},
	}
	return nil
}
```
