# Scenario

**Feature**: open a directory or a file in Visual Studio Code

```
# L2 harness (parallel-safe)
leaf Setup -> Operation + Kind + window mode / line + Home fixture
root Setup -> WorkDir + Dir and File fixtures under t.TempDir
root Run   -> Args / FileArgs / WellKnownCandidates (pure)
              OpenConfig / OpenFileConfig with injected Resolve + Run
leaf Assert-> argv, Result.CodePath/Via, resolve + runner spies, errors
```

## Preconditions

- Package under test:
  `github.com/xhd2015/dot-pkgs/go-pkgs/shell/vscode`
- All leaves are L2: injectable `Resolve` and `Run`. No real `code`, no VS Code
  window, no login-shell probe.
- Parallel-safe isolation: `t.TempDir()` only; no `os.Setenv` / `t.Setenv` /
  `os.Chdir` / `t.Chdir`.
- Out of suite: the Marcus daemon endpoint, the ai-workshop button,
  `shell/open`, `shell/openterm2`.

## Steps

1. Root `Setup` allocates `WorkDir`, a `Dir` fixture, a `File` fixture, and a
   `Home` directory, and zeros the inject fixtures.
2. Grouping `Setup` sets `Operation` (`args`, `candidates`, or `open`).
3. Leaf `Setup` sets `Kind`, window mode / line, and the failure injection.
4. Root `Run` dispatches to the matching pure builder or `*Config` entry.
5. Leaf `Assert` checks argv, `Result`, resolve/runner spies, and errors.

## Context

- Validation runs before resolution: an empty dir or file never resolves and
  never runs.
- Resolution happens exactly once, with `Config.Home`.
- The runner receives `args[0]` = the resolved CLI.
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
	req.Dir = filepath.Join(req.WorkDir, "worktrees", "credit-pricing-center", "master")
	if err := os.MkdirAll(req.Dir, 0o755); err != nil {
		return err
	}
	req.Home = filepath.Join(req.WorkDir, "home")
	if err := os.MkdirAll(req.Home, 0o755); err != nil {
		return err
	}
	req.File = filepath.Join(req.Dir, "service.go")
	if err := os.WriteFile(req.File, []byte("package main\n"), 0o644); err != nil {
		return err
	}
	req.Operation = ""
	req.Kind = ""
	req.Line = 0
	req.NewWindow = false
	req.CodePath = "/opt/code/bin/code"
	req.Via = ""
	req.ResolveErr = ""
	req.RunnerErr = ""
	req.RunOut = ""
	return nil
}
```
