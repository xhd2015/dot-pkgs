# Scenario

**Feature**: git-initialized entry reports one git repo aggregate

```
Scan -> with-git dir -> git/scan_repo -> Aggregates.GitRepos == 1
```

## Preconditions

`git` must be on PATH; test skips otherwise.

## Steps

1. Seed `git-dirs` profile: `with-git` git repo + plain `plain-dir`.
2. Set `req.Home` to temp dir.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if !gitAvailable(t) {
		return nil
	}
	home := t.TempDir()
	req.Home = home
	req.SeedProfile = "git-dirs"
	seedHome(t, home, req.SeedProfile)
	return nil
}
```