# Scenario

**Feature**: scan results appear in alphabetical order by entry name

```
Scan -> sorted HOME children -> []EntryResult in name order
```

## Steps

1. Seed `entry-order` profile: five mixed dir/file entries.
2. Set `req.Home` to temp dir.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	home := t.TempDir()
	req.Home = home
	req.SeedProfile = "entry-order"
	seedHome(t, home, req.SeedProfile)
	return nil
}
```