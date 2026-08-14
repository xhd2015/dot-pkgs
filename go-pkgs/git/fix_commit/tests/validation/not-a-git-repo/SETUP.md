# Scenario

**Feature**: `-C` directory is not a git repository

```
RunCLI -C <plain-dir> HEAD -m x -> Error: not a git repository: <dir>
```

## Steps

1. Create a temp directory that is not a git repo.
2. Set `req.Args` to `["-C", dir, "HEAD", "-m", "x"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Dir = t.TempDir()
	req.Args = []string{"-C", req.Dir, "HEAD", "-m", "x"}
	return nil
}
```
