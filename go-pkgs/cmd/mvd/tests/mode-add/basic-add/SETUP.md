# Scenario

Add a directory to tracking.

mvd --add tracked → [(tracked)]

## Steps
- Create a directory to track.
- Use --add to register it with mvd.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	dir := filepath.Join(req.WorkRoot, "tracked")
	mkdirAll(t, dir)
	req.Args = []string{"--add", dir}
	return nil
}
```
