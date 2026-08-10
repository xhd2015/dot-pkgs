# Scenario

**Feature**: batch-resolve CLI names under thin GUI PATH; derive PATH dirs

```
# L2 harness (parallel-safe)
leaf Setup -> Operation + Names + ExtraDirs/Home + inject fixtures
root Setup  -> WorkDir under t.TempDir (no t.Setenv / t.Chdir)
root Run    -> lookpath.LookupPaths / LookupItems.Dirs / DirsEnv
               with injectable LookPath / IsExecutable / RunLogin
leaf Assert -> LookupItems + From + Dirs/DirsEnv + error
```

## Preconditions

- Package under test:
  `github.com/xhd2015/dot-pkgs/go-pkgs/shell/lookpath`
- **Classic TDD:** public symbols under test do **not** exist yet; root `Run`
  imports them so the suite is **RED** (compile fail) until implementer lands:
  - `LookupPaths`, `LookupItem`, `LookupItems`
  - `(LookupItems).Dirs`, `(LookupItems).DirsEnv`
  - Reuse existing `Options` injectables
- All default leaves are L2: injectable LookPath / IsExecutable / RunLogin.
  No real login shell, no process PATH mutation.
- Parallel-safe isolation: `t.TempDir()` only; no `os.Setenv` / `t.Setenv` /
  `os.Chdir` / `t.Chdir` for config.
- Sibling sealed suite `./tests/lookpath/` is out of scope (do not modify).

## Steps

1. Root `Setup` allocates `WorkDir` and zeros inject fixtures.
2. Grouping `Setup` sets `Operation` (and stage-specific defaults).
3. Leaf `Setup` creates fixtures under `WorkDir` and fills Request fields.
4. Root `Run` dispatches to the public API with injectables.
5. Leaf `Assert` checks items / Dirs / DirsEnv / error and optional spies.

## Context

- Per-name order: path → extra_dir → default_dir → candidate → login → missing.
- `From` is `"bash"` | `"zsh"` only when login wins; otherwise `""`.
- `Dirs` unique first-seen parent dirs of found paths; `DirsEnv` joins with
  `os.PathListSeparator`.
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
	req.Names = nil
	req.Home = ""
	req.ExtraDirs = nil
	req.ExtraCandidates = nil
	req.Shells = nil
	req.Timeout = 0
	req.LookPathHits = nil
	req.ExecOverride = nil
	req.LoginStdout = nil
	req.LoginStdoutByName = nil
	req.LoginFail = nil
	return nil
}
```
