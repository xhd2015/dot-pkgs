# Scenario

Error when the directory does not exist.

mvd --add no-such-dir → error → does not exist

## Steps
- Attempt to add a directory path that does not exist on the filesystem.
- This should fail because `mvd --add` requires the target directory to exist.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"--add", filepath.Join(req.WorkRoot, "no-such-dir")}
	return nil
}
```
