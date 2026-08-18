# Scenario

Move back one step to the previous location.

mvd src dst → [(src), (dst/src)]
mvd --back dst/src → [(src)]

## Steps
- Create a source directory `src` containing a file and a destination directory `dst`.
- Move `src` into `dst` using `mvd src dst` — this renames `dst` to `dst` and creates the history entry with `src` as the root and `dst` as the current location.
- Run `mvd --back <moved-path>` to move the project back to its original location at `src`.

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
    
    req.Args = []string{src, dst}
    resp, err := runMvd(t, d, req)
    if err != nil { return err }
    if resp.ExitCode != 0 { t.Fatalf("move: %s", resp.Output) }
    
    p := filepath.Join(dst, "src")
    req.Args = []string{"--back", p}
    return nil
}
```
