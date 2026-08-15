# Scenario

**Feature**: top-level files report text line counts or binary marker

```
Scan -> file entry -> size + lines (text) or lines (binary)
```

## Steps

1. Seed `file-lines` profile from leaf `testdata/`.
2. Set `req.Home` to temp dir.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	home := t.TempDir()
	req.Home = home
	req.SeedProfile = "file-lines"
	seedHome(t, home, req.SeedProfile)
	return nil
}
```