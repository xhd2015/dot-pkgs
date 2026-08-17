# Scenario

**Feature**: an existing dest directory is returned; Install and Prompt are unused

```
# $InstallDir/go1.19.13 already a directory
dest dir present -> ResolveGoroot -> dest; Install not called; Prompt not written
```

## Steps

1. Create `$InstallDir/go1.19.13` as a directory.
2. Set a Prompt that must not appear on Stderr (install will not run).

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	dest := filepath.Join(req.InstallDir, "go1.19.13")
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}
	req.Prompt = "should-not-print\n"
	return nil
}
```
