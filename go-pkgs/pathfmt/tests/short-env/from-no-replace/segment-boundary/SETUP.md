# Scenario

**Feature**: env prefix match requires a path-segment boundary

```
X=/foo/ba
path=/foo/bar
-> no $X (TildeHome or absolute)
```

## Steps

1. Inject env `X=/foo/ba` (string prefix of path but not a segment boundary).
2. Set path to `/foo/bar`.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	// Synthetic absolute paths; need not exist on disk.
	req.Env = []string{envPair("X", filepath.Join(string(filepath.Separator)+"foo", "ba"))}
	req.Path = filepath.Join(string(filepath.Separator)+"foo", "bar")
	return nil
}
```
