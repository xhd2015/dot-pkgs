# Scenario

--dry-run with --rm -f: prints intent, history entry retained.

mvd --add tracked; mvd tracked dst → [(tracked), (dst/tracked)]
mvd --dry-run --rm -f tracked → prints 'would remove'

## Steps
- Add a directory to history, then move it to create movement history (2+ locations).
- Dry-run force-remove it with `--rm -f`.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	src := filepath.Join(req.WorkRoot, "src")
	dst := filepath.Join(req.WorkRoot, "dst")
	mkdirAll(t, src)
	mkdirAll(t, dst)
	// Move (creates history with 2 locations)
	req.Args = []string{src, dst}
	resp, err := runMvd(t, req)
	if err != nil { return err }
	if resp.ExitCode != 0 { t.Fatalf("move: %s", resp.Output) }

	// Now dry-run force remove
	req.Args = []string{"--dry-run", "--rm", "-f", src}
	return nil
}
```
