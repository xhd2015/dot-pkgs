# Scenario

Clear history using basename resolution.

mvd --add tracked → [(tracked)]
mvd tracked dst → [(tracked), (dst/tracked)]
mvd --clear tracked → []

## Steps
- Add a project directory to mvd.
- Change to a separate working directory to test basename lookup.
- Clear the project history using its basename.

```go
import (
    "os"
    "path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    projectRoot := filepath.Join(req.WorkRoot, "projects")
    dir := filepath.Join(projectRoot, "myproject")
    mkdirAll(t, dir)

    req.Args = []string{"--add", dir}
    resp, err := runMvd(t, d, req)
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

    req.Args = []string{"--clear", "myproject"}
    return nil
}
```
