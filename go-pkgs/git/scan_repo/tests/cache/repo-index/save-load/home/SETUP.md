# Scenario

**Feature**: round-trip repo index for universe `home`

```
# home universe
SaveRepoIndex(universe=home) -> <CacheRoot>/home/repos.json
  -> LoadRepoIndex(..., "home") returns same fields
```

## Steps

1. Set `Universe` to `"home"`.
2. Populate `req.Index` with v1 non-default values and one repo entry.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Universe = "home"
	req.Index = scan_repo.RepoIndex{
		Version:   1,
		Universe:  "home",
		Base:      "/Users/xhd2015",
		UpdatedAt: "2026-07-15T12:00:00Z",
		Repos: []scan_repo.RepoIndexEntry{
			{
				Path:     "/Users/xhd2015/Projects/org/alpha",
				RepoType: "main",
				GitDir:   "/Users/xhd2015/Projects/org/alpha/.git",
				Depth:    2,
				SeenAt:   "2026-07-15T11:30:00Z",
			},
		},
	}
	return nil
}
```
