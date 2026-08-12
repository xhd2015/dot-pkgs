# Scenario

**Feature**: resolve login interactive shell environment (full dump or single var)

```
# L2 harness (parallel-safe)
leaf Setup -> Operation + Shell + EnvName + NUL dump / LoginFail
root Setup  -> WorkDir under t.TempDir (no t.Setenv / t.Chdir)
root Run    -> lookpath.Resolve{Bash,Zsh}LoginEnv(s)
               with injectable LoginEnvOptions.RunLogin
leaf Assert -> Envs[] or Value + error + RunLogin shell spy
```

## Preconditions

- Package under test:
  `github.com/xhd2015/dot-pkgs/go-pkgs/shell/lookpath`
- **Classic TDD:** public symbols under test do **not** exist yet; root `Run`
  imports them so the suite is **RED** (compile fail) until implementer lands:
  - `LoginEnvOptions`
  - `ResolveBashLoginEnvs`, `ResolveZshLoginEnvs`
  - `ResolveBashLoginEnv`, `ResolveZshLoginEnv`
- All default leaves are L2: injectable `RunLogin` only. No real login shell,
  no process PATH / env mutation.
- Parallel-safe isolation: `t.TempDir()` only; no `os.Setenv` / `t.Setenv` /
  `os.Chdir` / `t.Chdir` for config.
- Sibling suites `./tests/lookpath/` and `./tests/lookup-paths/` are out of
  scope (do not modify).

## Steps

1. Root `Setup` allocates `WorkDir` and zeros inject fixtures.
2. Grouping `Setup` sets `Operation` and/or `Shell`.
3. Leaf `Setup` fills dump fixtures / EnvName / LoginFail.
4. Root `Run` dispatches to the public API with injectable `RunLogin`.
5. Leaf `Assert` checks Envs / Value / error and optional RunLogin spies.

## Context

- Production dump: `env -0` (NUL-delimited `KEY=value`); harness injects the
  same wire format via `LoginStdout`.
- Single-env: empty name → error; unset/empty → `("", nil)`; run fail → error.
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
	req.Shell = ""
	req.Home = ""
	req.Timeout = 0
	req.ShellBin = ""
	req.EnvName = ""
	req.LoginStdout = ""
	req.LoginFail = false
	return nil
}
```
