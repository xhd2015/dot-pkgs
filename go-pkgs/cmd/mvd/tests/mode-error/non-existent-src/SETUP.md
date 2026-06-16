# Scenario

Error when the source path does not exist.

mvd no-such-dir dst → error → does not exist

## Steps
- Try to move a non-existent source directory.
- The destination exists.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	nonexistent := filepath.Join(req.WorkRoot, "no-such-dir")
	dst := filepath.Join(req.WorkRoot, "dst")
	mkdirAll(t, dst)
	req.Args = []string{nonexistent, dst}
	return nil
}
```
