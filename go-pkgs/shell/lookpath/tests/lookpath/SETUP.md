# Scenario

**Feature**: resolve CLI binary names under thin GUI PATH (Launch Services)

```
# L2 harness (parallel-safe)
leaf Setup -> Operation + name + ExtraDirs/Home/candidates + inject fixtures
root Setup  -> WorkDir under t.TempDir (no t.Setenv / t.Chdir)
root Run    -> lookpath.Look / LookPath / DefaultDirs / IsExecutable
               with injectable LookPath / IsExecutable / RunLogin
leaf Assert -> Result.Path + Via (or pure helper outputs) + error text
```

## Preconditions

- Package under test:
  `github.com/xhd2015/dot-pkgs/go-pkgs/shell/lookpath`
- **Classic TDD:** package and public symbols do **not** exist yet; root `Run`
  imports them so the suite is **RED** (compile fail) until implementer lands:
  - `Look`, `LookPath`, `DefaultDirs`, `IsExecutable`
  - `Result`, `Options` (injectables `LookPath`, `IsExecutable`, `RunLogin`)
- All default leaves are L2: injectable LookPath / IsExecutable / RunLogin.
  No real login shell, no process PATH mutation.
- Parallel-safe isolation: `t.TempDir()` only; no `os.Setenv` / `t.Setenv` /
  `os.Chdir` / `t.Chdir` for config.
- Out of suite: Marcus wiring, agent-pro migration, real bash `-lic` e2e.

## Steps

1. Root `Setup` allocates `WorkDir` and zeros inject fixtures.
2. Grouping `Setup` sets `Operation` (and stage-specific defaults).
3. Leaf `Setup` creates fixtures under `WorkDir` and fills Request fields.
4. Root `Run` dispatches to the public API with injectables.
5. Leaf `Assert` checks Path/Via/error and optional LookPath call spies.

## Context

- Resolution order: direct → path → extra_dir → default_dir → candidate →
  login_shell → error.
- `Result.Via` values: `direct`, `path`, `extra_dir`, `default_dir`,
  `candidate`, `login_shell:bash`, `login_shell:zsh`, …
- DefaultDirs when home set: `$HOME/.local/bin`, `$HOME/go/bin`,
  `/opt/homebrew/bin`, `/usr/local/bin`.
- Product must not mutate process env/cwd.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.WorkDir = t.TempDir()
	req.Operation = ""
	req.Name = ""
	req.Home = ""
	req.ExtraDirs = nil
	req.ExtraCandidates = nil
	req.Shells = nil
	req.Timeout = 0
	req.LookPathHit = ""
	req.ExecOverride = nil
	req.LoginStdout = nil
	req.LoginFail = nil
	req.ExpectNoLookPath = false
	req.DefaultDirsHome = ""
	req.IsExecPath = ""
	return nil
}
```
