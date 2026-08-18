# Scenario

Move back using a registered alias.

mvd --add src; mvd --add-alias src al → [(src)]
mvd al dst → [(src), (dst/src)]
mvd --back al → [(src)]

## Steps
- Create a project whose root basename is `opencode-latest` (NOT `opencode`).
- Register an alias `opencode` pointing to this project via `--add-alias opencode opencode-latest`.
- Move the project by its alias `opencode` into `scratch`.
- Run `mvd --back opencode` from a separate working directory — this must resolve the alias to find the project in history.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    projectRoot := filepath.Join(req.WorkRoot, "projects")
    src := filepath.Join(projectRoot, "opencode-latest")
    dst := filepath.Join(req.WorkRoot, "scratch")
    mkdirAll(t, src)
    mkdirAll(t, dst)
    
    req.Args = []string{"--add", src}
    resp, err := runMvd(t, d, req)
    if err != nil { return err }
    if resp.ExitCode != 0 { t.Fatalf("add: %s", resp.Output) }
    
    req.Args = []string{"--add-alias", "opencode", "opencode-latest"}
    resp, err = runMvd(t, d, req)
    if err != nil { return err }
    if resp.ExitCode != 0 { t.Fatalf("add-alias: %s", resp.Output) }
    
    req.Args = []string{"opencode", dst}
    resp, err = runMvd(t, d, req)
    if err != nil { return err }
    if resp.ExitCode != 0 { t.Fatalf("move by alias: %s", resp.Output) }
    
    cwd := filepath.Join(req.WorkRoot, "cwd")
    mkdirAll(t, cwd)
	req.Cwd = cwd
    
    req.Args = []string{"--back", "opencode"}
    return nil
}
```
