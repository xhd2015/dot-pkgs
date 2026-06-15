## Steps
- Add some history, then run `mvd --dry-run --list` (read-only command).
- `--dry-run` should NOT affect read-only commands — they run normally.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	dir := filepath.Join(req.WorkRoot, "myproject")
	mkdirAll(t, dir)
	writeFile(t, filepath.Join(dir, "README.md"), "# test")
	// Add to history
	req.Args = []string{"--add", dir}
	resp, err := runMvd(t, req)
	if err != nil { return err }
	if resp.ExitCode != 0 { t.Fatalf("add: %s", resp.Output) }

	// Now dry-run list
	req.Args = []string{"--dry-run", "--list"}
	return nil
}
```
