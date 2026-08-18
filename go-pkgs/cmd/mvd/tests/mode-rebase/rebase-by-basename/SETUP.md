# Scenario

Rebase using basename resolution.

mvd src d1 → [(src), (d1/src)]
mvd --rebase src rebased → [(rebased), (d1/src)]

## Steps
- Add a project directory to mvd under a dedicated root.
- Change to a separate working directory to test basename lookup.
- Rebase the project using its basename.

```go
import (
    "path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    projectRoot := filepath.Join(req.WorkRoot, "projects")
    dir := filepath.Join(projectRoot, "myproject")
    newBase := filepath.Join(req.WorkRoot, "newbase")
    mkdirAll(t, dir)
    mkdirAll(t, newBase)

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
	req.Cwd = cwd

    req.Args = []string{"--rebase", "myproject", newBase}
    return nil
}
```
