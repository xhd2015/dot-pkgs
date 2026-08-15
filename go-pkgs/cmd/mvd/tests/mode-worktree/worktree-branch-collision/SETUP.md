# Scenario

Branch name collision generates a date-suffixed name.

mvd -w repo wt → [(repo), (wt w:wt)]
mvd -w repo wt → [(repo), (wt w:wt), (wt w:wt-YYYY-MM-DD)]

## Steps
- Create a git repo with an existing branch "myfeature".
- Use -w with a target dir that would normally produce branch "myfeature".

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "main")
	mkdirAll(t, mainRepo)
	initGitRepo(t, mainRepo)

	runGit(t, mainRepo, "branch", "myfeature")

	wtDir := filepath.Join(req.WorkRoot, "myfeature")
	req.Args = []string{"-w", mainRepo, wtDir}
	return nil
}
```
