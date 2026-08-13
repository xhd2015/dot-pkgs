# Scenario

**Feature**: a file path is not an openable directory

```
dir=regular file -> OpenConfig -> error
neither opener called
```

## Steps

1. Create a regular file under `WorkDir`.
2. Set `Dir` to that file path.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	p := filepath.Join(req.WorkDir, "not-a-dir.txt")
	if err := os.WriteFile(p, []byte("x\n"), 0o644); err != nil {
		return err
	}
	req.Dir = p
	return nil
}
```
