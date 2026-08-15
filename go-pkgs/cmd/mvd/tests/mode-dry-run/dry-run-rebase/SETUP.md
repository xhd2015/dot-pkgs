# Scenario

--dry-run with --rebase: prints intent, history unchanged.

mvd --add tracked; mvd tracked dst → [(tracked), (dst/tracked)]
mvd --dry-run --rebase tracked new → prints 'would rebase'

## Steps
- Add a directory to history, then dry-run rebase it to a new directory.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	dir := filepath.Join(req.WorkRoot, "oldbase")
	newDir := filepath.Join(req.WorkRoot, "newbase")
	mkdirAll(t, dir)
	mkdirAll(t, newDir)
	// First add
	req.Args = []string{"--add", dir}
	resp, err := runMvd(t, d, req)
	if err != nil { return err }
	if resp.ExitCode != 0 { t.Fatalf("add: %s", resp.Output) }

	// Now dry-run rebase
	req.Args = []string{"--dry-run", "--rebase", dir, newDir}
	return nil
}
```
