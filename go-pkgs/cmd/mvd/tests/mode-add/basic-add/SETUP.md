# Scenario

Add a directory to tracking.

mvd --add tracked → [(tracked)]

## Steps
- Create a directory to track.
- Use --add to register it with mvd.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	dir := filepath.Join(req.WorkRoot, "tracked")
	mkdirAll(t, dir)
	req.Args = []string{"--add", dir}
	return nil
}
```
