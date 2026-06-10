## Steps
- Create a git repo with an existing branch "myfeature".
- Use -w with a target dir that would normally produce branch "myfeature".

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "main")
	mkdirAll(t, mainRepo)
	initGitRepo(t, mainRepo)

	runGit(t, mainRepo, "branch", "myfeature")

	wtDir := filepath.Join(req.WorkRoot, "myfeature")
	req.Args = []string{"-w", mainRepo, wtDir}
	return nil
}
```
