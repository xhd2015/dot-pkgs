# Scenario

--dry-run move into existing dir: basename join, no actual move.

mvd --dry-run src existing-dir → prints 'would move'  (no change)

## Steps
- Create a source directory `src` and an existing destination directory `dst`.
- Run `mvd --dry-run src dst` where `dst` is an existing directory.
- The source should be moved *into* `dst` (basename join), so the target becomes `dst/src`.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	src := filepath.Join(req.WorkRoot, "src")
	dst := filepath.Join(req.WorkRoot, "dst")
	mkdirAll(t, src)
	mkdirAll(t, dst)
	writeFile(t, filepath.Join(src, "f.txt"), "hello")
	req.Args = []string{"--dry-run", src, dst}
	return nil
}
```
