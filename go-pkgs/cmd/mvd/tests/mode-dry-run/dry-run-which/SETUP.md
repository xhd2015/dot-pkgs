# Scenario

--dry-run with --which: read-only, runs normally.

mvd --add tracked → [(tracked)]
mvd --dry-run --which tracked → prints path normally

## Steps
- Add some history, then run `mvd --dry-run --which proj` (read-only command).
- `--dry-run` should NOT affect read-only commands.

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

	// Now dry-run which
	req.Args = []string{"--dry-run", "--which", dir}
	return nil
}
```
