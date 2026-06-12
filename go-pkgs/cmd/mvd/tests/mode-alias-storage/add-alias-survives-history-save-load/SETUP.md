## Steps
- Create two project directories and add both to mvd tracking.
- Register an alias for the first project.
- Move the second project (triggers history save/load cycle).
- Verify alias survived by running `--which` on the alias.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	projDir := filepath.Join(req.WorkRoot, "projects", "myproj")
	otherDir := filepath.Join(req.WorkRoot, "projects", "other")
	scratch := filepath.Join(req.WorkRoot, "scratch")
	mkdirAll(t, projDir)
	mkdirAll(t, otherDir)
	mkdirAll(t, scratch)

	req.Args = []string{"--add", projDir}
	resp, err := runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("add proj: %s", resp.Output)
	}

	req.Args = []string{"--add", otherDir}
	resp, err = runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("add other: %s", resp.Output)
	}

	req.Args = []string{"--add-alias", "mp", "myproj"}
	resp, err = runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("add-alias: %s", resp.Output)
	}

	req.Args = []string{otherDir, scratch}
	resp, err = runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("move other: %s", resp.Output)
	}

	req.Args = []string{"--which", "mp"}
	return nil
}
```
