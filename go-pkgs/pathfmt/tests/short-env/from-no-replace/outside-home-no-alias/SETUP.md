# Scenario

**Feature**: path outside home with no matching alias stays absolute

```
env empty / unrelated
path under temp (outside home) -> absolute
```

## Steps

1. Create a temp path outside home; empty env.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	outside := filepath.Join(os.TempDir(), "doctest-short-env-outside-home")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		return err
	}
	req.Env = []string{}
	req.Path = outside
	return nil
}
```
