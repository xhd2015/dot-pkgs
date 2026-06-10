## Steps
- Move src to dst so it has a history entry.
- Then clear the history for src.

```go
import (
    "path/filepath"
)

func Setup(t *testing.T, req *Request) error {
    src := filepath.Join(req.WorkRoot, "src")
    dst := filepath.Join(req.WorkRoot, "dst")
    mkdirAll(t, src)
    mkdirAll(t, dst)

    req.Args = []string{src, dst}
    resp, err := runMvd(t, req)
    if err != nil {
        return err
    }
    if resp.ExitCode != 0 {
        t.Fatalf("move: %s", resp.Output)
    }

    req.Args = []string{"--clear", src}
    return nil
}
```
