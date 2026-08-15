# Scenario

--dry-run with --add-alias: prints intent, alias not persisted.

mvd --add tracked → [(tracked)]
mvd --dry-run --add-alias tracked al → prints 'would add alias'

## Steps
- Add a project to the history first (using normal `mvd --add`).
- Then run `mvd --dry-run --add-alias myalias PROJECT` to dry-run adding an alias.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	dir := filepath.Join(req.WorkRoot, "myproject")
	mkdirAll(t, dir)
	// First add the project normally
	req.Args = []string{"--add", dir}
	resp, err := runMvd(t, d, req)
	if err != nil { return err }
	if resp.ExitCode != 0 { t.Fatalf("add: %s", resp.Output) }

	// Now dry-run add-alias
	req.Args = []string{"--dry-run", "--add-alias", "myalias", dir}
	return nil
}
```
