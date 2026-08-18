# Scenario

**Feature**: `.codex` entry gets rollout sessions, skill dirs, and zero plugins

```
Scan -> .codex dir -> semantic lines (sessions rollouts, skills, plugins 0)
```

## Steps

1. Seed `codex` profile: two rollout files, one skill dir, no plugins dir.
2. Set `req.Home` to temp dir.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	home := t.TempDir()
	req.Home = home
	req.SeedProfile = "codex"
	seedHome(t, d, home, req.SeedProfile)
	return nil
}
```