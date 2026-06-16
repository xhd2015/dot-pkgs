# Scenario

--clear with $X/myproject via lls env var expansion.

mvd --add $X/myproject; mvd $X/myproject dst → [(projects/myproject), (dst/myproject)]
mvd --clear $X/myproject → []

## Steps
- Set up lls config with X env var.
- Create projects/myproject directory, add it.
- --clear using $X/myproject.

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

	req.Args = []string{"--clear", "$X/myproject"}
	return nil
}
```
