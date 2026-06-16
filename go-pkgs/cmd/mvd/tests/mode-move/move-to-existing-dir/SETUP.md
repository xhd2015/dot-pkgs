# Scenario

Move into an existing directory — mvd joins the basename.

mvd mysrc existing-dir → [(mysrc), (existing-dir/mysrc)]

## Steps
- Create a source directory (mysrc) with a test file.
- Create an existing destination directory.
- Set req.Args to move the source into the existing destination.
- mvd should join the destination with the basename of the source.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	src := filepath.Join(req.WorkRoot, "mysrc")
	dst := filepath.Join(req.WorkRoot, "existing-dir")
	mkdirAll(t, src)
	mkdirAll(t, dst)
	writeFile(t, filepath.Join(src, "f.txt"), "hello")
	req.Args = []string{src, dst}
	return nil
}
```
