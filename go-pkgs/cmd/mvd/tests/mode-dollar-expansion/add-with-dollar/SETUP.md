# Scenario

--add with $X/myproject via lls env var expansion.

mvd --add $X/myproject → [(projects/myproject)]

## Steps
- Set up lls config with X env var.
- Create projects/myproject directory.
- Run --add with $X/myproject.

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
	return nil
}
```
