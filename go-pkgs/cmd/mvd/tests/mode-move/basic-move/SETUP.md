# Scenario

Plain move of a non-git directory.

mvd src dst → [(src), (dst)]

## Steps
- Create src and dst directories under WorkRoot.
- Write a test file in src.
- Set req.Args to move src to dst using absolute paths.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
    src := filepath.Join(req.WorkRoot, "src")
    dst := filepath.Join(req.WorkRoot, "dst")
    mkdirAll(t, src)
    writeFile(t, filepath.Join(src, "f.txt"), "hello")
    req.Args = []string{src, dst}
	return nil
}
```
