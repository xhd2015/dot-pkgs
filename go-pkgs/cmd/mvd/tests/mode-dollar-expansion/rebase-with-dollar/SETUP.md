## Steps
- Set up lls config with X env var.
- Create projects/myproject directory, add it, move to d1.
- --rebase $X/myproject onto newBase.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	homeDir := filepath.Join(req.WorkRoot, ".lls-home")
	configDir := filepath.Join(homeDir, "Library", "Application Support", "lls")
	mkdirAll(t, configDir)
	writeFile(t, filepath.Join(configDir, "config.json"), `{"envs":["X"]}`)

	projectRoot := filepath.Join(req.WorkRoot, "projects")
	dir := filepath.Join(projectRoot, "myproject")
	mkdirAll(t, dir)

	req.Args = []string{"--add", "$X/myproject"}
	resp, err := runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("add: %s", resp.Output)
	}

	d1 := filepath.Join(req.WorkRoot, "d1")
	mkdirAll(t, d1)
	req.Args = []string{"$X/myproject", d1}
	resp, err = runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("move: %s", resp.Output)
	}

	newBase := filepath.Join(req.WorkRoot, "rebased")
	mkdirAll(t, newBase)
	req.Args = []string{"--rebase", "$X/myproject", newBase}
	return nil
}
```
