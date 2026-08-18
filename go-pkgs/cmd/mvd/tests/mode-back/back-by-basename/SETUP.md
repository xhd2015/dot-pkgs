# Scenario

Move back using a unique root basename.

mvd src dst → [(src), (dst/src)]
mvd --back src → [(src)]

## Steps
- Create a project under `projects/kool` and add it to mvd's tracking via `--add`.
- Move the project to a `scratch` directory using its unique basename `"kool"` from a different working directory.
- Run `mvd --back kool` from the separate working directory — the basename lookup should find the project even though the CWD is unrelated.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    projectRoot := filepath.Join(req.WorkRoot, "projects")
    src := filepath.Join(projectRoot, "kool")
    dst := filepath.Join(req.WorkRoot, "scratch")
    mkdirAll(t, src)
    mkdirAll(t, dst)
    
    req.Args = []string{"--add", src}
    resp, err := runMvd(t, d, req)
    if err != nil { return err }
    if resp.ExitCode != 0 { t.Fatalf("add: %s", resp.Output) }
    
    req.Args = []string{"kool", dst}
    resp, err = runMvd(t, d, req)
    if err != nil { return err }
    if resp.ExitCode != 0 { t.Fatalf("move by basename: %s", resp.Output) }
    
    cwd := filepath.Join(req.WorkRoot, "cwd")
    mkdirAll(t, cwd)
	req.Cwd = cwd
    
    req.Args = []string{"--back", "kool"}
    return nil
}
```
