# Scenario

--dry-run --back at origin: no-op, no dry-run message.

mvd --dry-run --back src → no-op → nothing to move back

## Steps
- Add a directory to history (single entry, at origin).
- Dry-run `--back` — should be a no-op (nothing to move back to).

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	dir := filepath.Join(req.WorkRoot, "myproject")
	mkdirAll(t, dir)
	writeFile(t, filepath.Join(dir, "f.txt"), "hello")
	// Add to history
	req.Args = []string{"--add", dir}
	resp, err := runMvd(t, req)
	if err != nil { return err }
	if resp.ExitCode != 0 { t.Fatalf("add: %s", resp.Output) }

	// Dry-run back at origin
	req.Args = []string{"--dry-run", "--back", dir}
	return nil
}
```
