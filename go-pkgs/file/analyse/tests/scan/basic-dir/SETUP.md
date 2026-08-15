# Scenario

**Feature**: directory entry reports sorted children and deep-aggregated sizes

```
Scan -> plain-dir EntryKindDir -> immediate children sorted by name
```

## Steps

1. Seed `basic` profile: `plain-dir/sub/nested.txt` and `notes.txt`.
2. Set `req.Home` to temp dir.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	home := t.TempDir()
	req.Home = home
	req.SeedProfile = "basic"
	seedHome(t, home, req.SeedProfile)
	return nil
}
```