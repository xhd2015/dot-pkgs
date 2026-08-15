# Scenario

**Feature**: replace under worktree/external/ is allowed in lenient hook mode

```
# scanning scope = current worktree toplevel
# replace => <worktree>/external/<dep-wt> is inside the worktree tree
# nested linked wt has its own git toplevel but must not block the hook
worktree + external linked wt + abs replace -> hook -> exit 0
```

## Preconditions

- Consumer git repo with linked external worktree under `external/mydep-main-2026-07-04`.
- Consumer `go.mod` has absolute replace to the external path (wrk style).
- Hook runs in default lenient mode (no `--strict`).

## Steps

1. Initialize consumer and dep repos; add linked external worktree.
2. Write absolute replace in consumer `go.mod`.
3. Run `git-hook-go-no-local-replace` from consumer repo root.

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

const wrkExternalHookDepModule = "example.com/dep"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = nil

	depMain := filepath.Join(req.RepoDir, "dep-main")
	externalPath := filepath.Join(req.RepoDir, "external", "mydep-main-2026-07-04")
	if err := os.MkdirAll(depMain, 0o755); err != nil {
		return err
	}
	if err := runGit(req.RepoDir, "init"); err != nil {
		return err
	}
	if err := runGit(depMain, "init"); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(depMain, "go.mod"), []byte("module "+wrkExternalHookDepModule+"\n\ngo 1.22\n"), 0o644); err != nil {
		return err
	}
	if err := runGit(depMain, "add", "go.mod"); err != nil {
		return err
	}
	if err := runGit(depMain, "commit", "-m", "add dep"); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(externalPath), 0o755); err != nil {
		return err
	}
	cmd := exec.Command("git", "-C", depMain, "worktree", "add", "-b", "mydep-main-2026-07-04", externalPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree add: %w: %s", err, strings.TrimSpace(string(out)))
	}

	goMod := "module example.com/consumer\n\ngo 1.22\n\nrequire " + wrkExternalHookDepModule + " v0.0.0\n\nreplace " + wrkExternalHookDepModule + " => " + externalPath + "\n"
	return os.WriteFile(filepath.Join(req.RepoDir, "go.mod"), []byte(goMod), 0o644)
}

```