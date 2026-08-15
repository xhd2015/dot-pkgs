# Scenario

Move using a registered alias for the source.

mvd --add src → [(src)]
mvd --add-alias src al → [(src)]
mvd al dst → [(src), (dst/src)]

## Steps
- Create a project at projects/kool.
- Move it to a scratch directory (first move).
- Create an alias "kk" for "kool".
- Change CWD to avoid local shadowing.
- Move the project by alias to a final directory.

```go
import (
	"os"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	projectRoot := filepath.Join(req.WorkRoot, "projects")
	src := filepath.Join(projectRoot, "kool")
	dst1 := filepath.Join(req.WorkRoot, "scratch")
	dst2 := filepath.Join(req.WorkRoot, "final")
	mkdirAll(t, src)
	mkdirAll(t, dst1)
	mkdirAll(t, dst2)

	req.Args = []string{src, dst1}
	resp, err := runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("first move: %s", resp.Output)
	}

	req.Args = []string{"--add-alias", "kk", "kool"}
	resp, err = runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("add-alias: %s", resp.Output)
	}

	cwd := filepath.Join(req.WorkRoot, "cwd")
	mkdirAll(t, cwd)
	if err := os.Chdir(cwd); err != nil {
		return err
	}

	req.Args = []string{"kk", dst2}
	return nil
}
```
