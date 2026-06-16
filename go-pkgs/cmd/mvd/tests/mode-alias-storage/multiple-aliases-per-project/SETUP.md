# Scenario

Multiple aliases for the same project stored in history.json.

mvd --add repo; --add-alias repo a1; --add-alias repo a2 → [(repo)]  (2 aliases)

## Steps
- Create a project directory and add it to mvd tracking.
- Register two aliases ("mp" and "myproj-alias") for the same project.
- Run `--list` to show all aliases in output.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	projDir := filepath.Join(req.WorkRoot, "projects", "myproj")
	mkdirAll(t, projDir)

	req.Args = []string{"--add", projDir}
	resp, err := runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("add: %s", resp.Output)
	}

	req.Args = []string{"--add-alias", "mp", "myproj"}
	resp, err = runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("add-alias mp: %s", resp.Output)
	}

	req.Args = []string{"--add-alias", "myproj-alias", "myproj"}
	resp, err = runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("add-alias myproj-alias: %s", resp.Output)
	}

	req.Args = []string{"--list"}
	return nil
}
```
