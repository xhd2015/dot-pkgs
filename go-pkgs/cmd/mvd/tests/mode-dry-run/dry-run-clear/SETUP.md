## Steps
- Move a directory to create history, then dry-run clear its history.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	src := filepath.Join(req.WorkRoot, "src")
	dst := filepath.Join(req.WorkRoot, "dst")
	mkdirAll(t, src)
	mkdirAll(t, dst)
	// First move to create history
	req.Args = []string{src, dst}
	resp, err := runMvd(t, req)
	if err != nil { return err }
	if resp.ExitCode != 0 { t.Fatalf("move: %s", resp.Output) }

	// Now dry-run clear
	req.Args = []string{"--dry-run", "--clear", src}
	return nil
}
```
