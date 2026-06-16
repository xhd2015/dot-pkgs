## Steps
- Set up lls config with X env var.
- Create projects/myproject with a file.
- Add it, then move with $X/myproject to dst.

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
	writeFile(t, filepath.Join(dir, "f.txt"), "hello")

	req.Args = []string{"--add", "$X/myproject"}
	resp, err := runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("add: %s", resp.Output)
	}

	dst := filepath.Join(req.WorkRoot, "dst")
	req.Args = []string{"$X/myproject", dst}
	return nil
}
```
