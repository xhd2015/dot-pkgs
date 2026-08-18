# Scenario

**Feature**: entry reports immediate `node_modules` child and recursive dir count

```
Scan -> nm-entry -> child node_modules + NodeModulesDirs == 2
```

## Steps

1. Seed `node-modules` profile: two nested `node_modules` directories.
2. Set `req.Home` to temp dir.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	home := t.TempDir()
	req.Home = home
	req.SeedProfile = "node-modules"
	seedHome(t, d, home, req.SeedProfile)
	return nil
}
```