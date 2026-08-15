# Scenario

Rebase a project entry to a new root path.

mvd src d1 → [(src), (d1/src)]
mvd --rebase src rebased → [(rebased), (d1/src)]

## Steps
- Move src to d1 so it has a history entry.
- Then rebase src onto newBase.

```go
import (
    "path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    src := filepath.Join(req.WorkRoot, "src")
    d1 := filepath.Join(req.WorkRoot, "d1")
    newBase := filepath.Join(req.WorkRoot, "rebased")
    mkdirAll(t, src)
    mkdirAll(t, d1)
    mkdirAll(t, newBase)

    req.Args = []string{src, d1}
    resp, err := runMvd(t, d, req)
    if err != nil {
        return err
    }
    if resp.ExitCode != 0 {
        t.Fatalf("move: %s", resp.Output)
    }

    req.Args = []string{"--rebase", src, newBase}
    return nil
}
```
