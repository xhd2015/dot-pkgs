# Scenario

--dry-run plain move: prints intent, skips actual move.

mvd --dry-run src dst → prints 'would move'  (no actual change)

## Steps
- Create a source directory `src` with a file under WorkRoot.
- Run `mvd --dry-run src dst` to exercise dry-run plain move.
- The destination `dst` does not exist yet, so it becomes the new name directly.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	src := filepath.Join(req.WorkRoot, "src")
	dst := filepath.Join(req.WorkRoot, "dst")
	mkdirAll(t, src)
	writeFile(t, filepath.Join(src, "f.txt"), "hello")
	req.Args = []string{"--dry-run", src, dst}
	return nil
}
```
