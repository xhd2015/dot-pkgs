# Scenario

--list with $X/myproject via lls env var expansion.

mvd --add $X/myproject → [(projects/myproject)]
mvd --list $X/myproject → shows chain

## Steps
- Set up lls config with X env var.
- Create projects/myproject directory, add it.
- --list using $X/myproject.

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

	req.Args = []string{"--add", "$X/myproject"}
	resp, err := runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("add: %s", resp.Output)
	}

	req.Args = []string{"--list", "$X/myproject"}
	return nil
}
```
