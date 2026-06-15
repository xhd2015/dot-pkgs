## Steps
- Add a directory to history, then dry-run rebase it to a new directory.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	dir := filepath.Join(req.WorkRoot, "oldbase")
	newDir := filepath.Join(req.WorkRoot, "newbase")
	mkdirAll(t, dir)
	mkdirAll(t, newDir)
	// First add
	req.Args = []string{"--add", dir}
	resp, err := runMvd(t, req)
	if err != nil { return err }
	if resp.ExitCode != 0 { t.Fatalf("add: %s", resp.Output) }

	// Now dry-run rebase
	req.Args = []string{"--dry-run", "--rebase", dir, newDir}
	return nil
}
```
