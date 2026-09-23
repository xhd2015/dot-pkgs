# Scenario

**Feature**: open a path, a URL, or a path with an application

```
# L2 harness (parallel-safe)
leaf Setup -> Operation + Kind + platform + fixtures
root Setup -> WorkDir + an existing target path under t.TempDir
root Run   -> Args/URLArgs/AppArgs (pure) or *Config (injected runner)
leaf Assert-> argv equality, Result.Out, runner spies, errors
```

## Preconditions

- Package under test:
  `github.com/xhd2015/dot-pkgs/go-pkgs/shell/open`
- All leaves are L2: injectable `GOOS` and `Run`. No real `open`,
  `xdg-open`, or `cmd start`.
- Parallel-safe isolation: `t.TempDir()` only; no `os.Setenv` / `t.Setenv` /
  `os.Chdir` / `t.Chdir`.
- Out of suite: `shell/openterm2`, `shell/vscode`, `open -R` reveal,
  bundle-id (`-b`) launches.

## Steps

1. Root `Setup` allocates `WorkDir` and an existing `Target` directory, and
   zeros the inject fixtures.
2. Grouping `Setup` sets `Operation` (`args` or `run`).
3. Leaf `Setup` sets `Kind` (path / url / app) and the fields that case needs.
4. Root `Run` dispatches to the matching pure builder or `*Config` entry.
5. Leaf `Assert` checks argv, `Result.Out`, runner call spies, and errors.

## Context

- Validation runs before the runner: an invalid target never launches.
- The windows argv carries an explicit empty window title before the target.
- `open -a` is darwin only; other platforms are rejected without launching.
- The product must not mutate process env or cwd.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.WorkDir = t.TempDir()
	req.Target = filepath.Join(req.WorkDir, "project")
	if err := os.MkdirAll(req.Target, 0o755); err != nil {
		return err
	}
	req.Operation = ""
	req.Kind = ""
	req.URL = ""
	req.AppName = ""
	req.AppPath = ""
	req.GOOS = ""
	req.Browser = ""
	req.NewWindow = false
	req.RunnerErr = ""
	return nil
}
```
