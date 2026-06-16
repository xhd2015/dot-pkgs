# Scenario

--dry-run with --rm: prints intent, history entry retained.

mvd --add tracked → [(tracked)]
mvd --dry-run --rm tracked → prints 'would remove'

## Steps
- Add a directory to history, then dry-run remove it.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	dir := filepath.Join(req.WorkRoot, "myproject")
	mkdirAll(t, dir)
	// First add the project
	req.Args = []string{"--add", dir}
	resp, err := runMvd(t, req)
	if err != nil { return err }
	if resp.ExitCode != 0 { t.Fatalf("add: %s", resp.Output) }

	// Now dry-run remove
	req.Args = []string{"--dry-run", "--rm", dir}
	return nil
}
```
