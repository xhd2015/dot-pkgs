# Scenario

Error when a basename matches multiple roots.

mvd --add a1/src → [(a1/src)]
mvd --add a2/src → [… (a2/src)]
mvd src dst → error → ambiguous

## Steps
- Create two projects with the same basename (kool) at different paths.
- Track both projects with --add.
- Change CWD to avoid local file shadowing.
- Try to move by basename, which should be ambiguous.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	first := filepath.Join(req.WorkRoot, "projects", "kool")
	second := filepath.Join(req.WorkRoot, "projects", "v2", "kool")
	dst := filepath.Join(req.WorkRoot, "dst")
	mkdirAll(t, first)
	mkdirAll(t, second)
	mkdirAll(t, dst)

	req.Args = []string{"--add", first}
	resp, err := runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("add first: %s", resp.Output)
	}

	req.Args = []string{"--add", second}
	resp, err = runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("add second: %s", resp.Output)
	}

	cwd := filepath.Join(req.WorkRoot, "cwd")
	mkdirAll(t, cwd)
	req.Cwd = cwd

	req.Args = []string{"kool", dst}
	return nil
}
```
