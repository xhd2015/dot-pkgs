# Scenario

Plain move with $X/myproject via lls env var expansion.

mvd --add $X/myproject → [(projects/myproject)]
mvd $X/myproject dst → [(projects/myproject), (dst/myproject)]

## Steps
- Set up lls config with X env var.
- Create projects/myproject with a file.
- Add it, then move with $X/myproject to dst.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	homeDir := filepath.Join(req.WorkRoot, ".lls-home")
	writeLlsXConfig(t, homeDir)

	projectRoot := filepath.Join(req.WorkRoot, "projects")
	dir := filepath.Join(projectRoot, "myproject")
	mkdirAll(t, dir)
	writeFile(t, filepath.Join(dir, "f.txt"), "hello")

	req.Args = []string{"--add", "$X/myproject"}
	resp, err := runMvd(t, d, req)
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
