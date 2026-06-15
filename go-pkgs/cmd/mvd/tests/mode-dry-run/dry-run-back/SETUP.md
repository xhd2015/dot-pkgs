## Steps
- Move a source directory to a destination, then dry-run `--back` from the moved location.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	src := filepath.Join(req.WorkRoot, "src")
	dst := filepath.Join(req.WorkRoot, "dst")
	mkdirAll(t, src)
	mkdirAll(t, dst)
	writeFile(t, filepath.Join(src, "f.txt"), "hello")
	// First move
	req.Args = []string{src, dst}
	resp, err := runMvd(t, req)
	if err != nil { return err }
	if resp.ExitCode != 0 { t.Fatalf("move: %s", resp.Output) }

	// Now dry-run back
	p := filepath.Join(dst, "src")
	req.Args = []string{"--dry-run", "--back", p}
	return nil
}
```
