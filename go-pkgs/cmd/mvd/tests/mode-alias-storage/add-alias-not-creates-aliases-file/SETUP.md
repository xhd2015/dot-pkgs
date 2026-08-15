# Scenario

--add-alias does not create aliases.json; alias stored in history.json.

mvd --add repo; mvd --add-alias repo al → [(repo)]  (alias in history.json)

## Steps
- Create a project directory.
- Add the project to mvd tracking via `--add`.
- Register an alias for the project via `--add-alias`.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	projDir := filepath.Join(req.WorkRoot, "projects", "myproj")
	mkdirAll(t, projDir)

	req.Args = []string{"--add", projDir}
	resp, err := runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("add: %s", resp.Output)
	}

	req.Args = []string{"--add-alias", "mp", "myproj"}
	resp, err = runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("add-alias: %s", resp.Output)
	}

	req.Args = []string{"--which", "mp"}
	return nil
}
```
