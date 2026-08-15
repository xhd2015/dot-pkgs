# Scenario

**Feature**: round-trip repo index for universe `root`

```
# root universe
SaveRepoIndex(universe=root) -> <CacheRoot>/root/repos.json
  -> LoadRepoIndex(..., "root") returns same fields
```

## Steps

1. Set `Universe` to `"root"`.
2. Populate `req.Index` with v1 non-default values distinct from the home leaf.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Universe = "root"
	req.Index = scan_repo.RepoIndex{
		Version:   1,
		Universe:  "root",
		Base:      "/",
		UpdatedAt: "2026-07-15T13:00:00Z",
		Repos: []scan_repo.RepoIndexEntry{
			{
				Path:     "/opt/src/beta",
				RepoType: "main",
				GitDir:   "/opt/src/beta/.git",
				Depth:    3,
				SeenAt:   "2026-07-15T12:45:00Z",
			},
			{
				Path:     "/opt/src/beta-wt",
				RepoType: "worktree",
				GitDir:   "/opt/src/beta/.git/worktrees/beta-wt",
				Depth:    3,
				SeenAt:   "2026-07-15T12:46:00Z",
			},
		},
	}
	return nil
}
```
