## Steps
- Set up lls config with X env var.
- Create projects/myproject directory.
- Add with $X/myproject, move to dst, then --back with $X/myproject.

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

	dst := filepath.Join(req.WorkRoot, "dst")
	mkdirAll(t, dst)
	req.Args = []string{"$X/myproject", dst}
	resp, err = runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("move: %s", resp.Output)
	}

	req.Args = []string{"--back", "$X/myproject"}
	return nil
}
```
