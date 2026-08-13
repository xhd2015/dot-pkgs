# Scenario

**Feature**: open a directory in iTerm2 when resolvable, else Terminal.app

```
# L2 harness (parallel-safe)
leaf Setup -> Operation + dir fixture + ResolveITerm / opener inject
root Setup  -> WorkDir + ValidDir under t.TempDir (no t.Setenv / t.Chdir)
root Run    -> openterm2.OpenConfig / TerminalOpenArgs
               with injectable ResolveITerm / OpenITerm / OpenTerminal
leaf Assert -> Result.Via + AppPath (or argv) + opener spies
```

## Preconditions

- Package under test:
  `github.com/xhd2015/dot-pkgs/go-pkgs/shell/openterm2`
- **Classic TDD:** package and public symbols do **not** exist yet; root `Run`
  imports them so the suite is **RED** (compile fail) until implementer lands:
  - `Open`, `OpenConfig`, `TerminalOpenArgs`
  - `Result`, `Config` (injectables `ResolveITerm`, `OpenITerm`, `OpenTerminal`)
  - `ViaITerm2`, `ViaTerminal`
- All default leaves are L2: injectable resolve + both openers. No real
  iTerm2, no real `open -a`, no `osascript`.
- Parallel-safe isolation: `t.TempDir()` only; no `os.Setenv` / `t.Setenv` /
  `os.Chdir` / `t.Chdir`.
- Out of suite: Marcus wiring, `shell/iterm2` contract changes,
  `ITERM2_APP_PATH` order.

## Steps

1. Root `Setup` allocates `WorkDir` and an existing `ValidDir`, and zeros inject fixtures.
2. Grouping `Setup` sets `Operation` (and resolve / app defaults).
3. Leaf `Setup` creates file/dir fixtures and fills Request fields.
4. Root `Run` dispatches to `OpenConfig` or `TerminalOpenArgs` with injectables.
5. Leaf `Assert` checks Via/AppPath/argv/error and opener call spies.

## Context

- Validation runs before either opener. Invalid dir never opens a terminal.
- `ResolveITerm()` non-empty → iTerm only; opener error does not fall through.
- `ResolveITerm()` empty → Terminal only; `AppPath` is `TerminalApp` or
  `/Applications/Utilities/Terminal.app`.
- `TerminalOpenArgs` is a pure argv helper; it does not exec.
- Product must not mutate process env/cwd.
- Openers receive `dir` as given (this suite uses absolute `t.TempDir` paths).

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
	req.ValidDir = filepath.Join(req.WorkDir, "project")
	if err := os.MkdirAll(req.ValidDir, 0o755); err != nil {
		return err
	}
	req.Operation = ""
	req.Dir = ""
	req.ITermAppPath = ""
	req.OpenITermErr = ""
	req.OpenTerminalErr = ""
	req.TerminalApp = ""
	req.ArgsAppPath = ""
	return nil
}
```
