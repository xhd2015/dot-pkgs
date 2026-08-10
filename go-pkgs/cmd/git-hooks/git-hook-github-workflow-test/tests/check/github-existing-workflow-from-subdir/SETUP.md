## Preconditions

- The repository origin is `github.com`.
- `.github/workflows/test.yml` already exists at the git repository root.
- The command runs from a subdirectory of the repository.

## Steps

1. Create a matching workflow file at the git repository root.
2. Create a `task-hub` subdirectory inside the repository.
3. Run the command in check mode from that subdirectory.

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CaseName = "check github existing workflow from subdir"
	fix := exec.Command(req.ToolPath, "--fix")
	fix.Dir = req.RepoDir
	if output, err := fix.CombinedOutput(); err != nil {
		return fmt.Errorf("bootstrap workflow with --fix: %w: %s", err, strings.TrimSpace(string(output)))
	}
	subdir := filepath.Join(req.RepoDir, "task-hub")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		return err
	}
	req.RunDir = subdir
	return nil
}
```