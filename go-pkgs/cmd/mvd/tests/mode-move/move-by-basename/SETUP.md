## Steps
- Create a project under WorkRoot/projects/kool.
- Use --add to track it.
- Change CWD to a different directory to avoid local file shadowing.
- Move the project by its basename to an archive directory.

```go
import (
	"os"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	projectRoot := filepath.Join(req.WorkRoot, "projects")
	src := filepath.Join(projectRoot, "kool")
	dst := filepath.Join(req.WorkRoot, "archive")
	mkdirAll(t, src)
	mkdirAll(t, dst)

	req.Args = []string{"--add", src}
	resp, err := runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("add: %s", resp.Output)
	}

	cwd := filepath.Join(req.WorkRoot, "cwd")
	mkdirAll(t, cwd)
	if err := os.Chdir(cwd); err != nil {
		return err
	}

	req.Args = []string{"kool", dst}
	return nil
}
```
