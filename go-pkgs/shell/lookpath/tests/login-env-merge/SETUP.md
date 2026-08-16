# Scenario

**Feature**: resolve login envs by detected shell; merge env slices last-wins

```
# L2 harness (parallel-safe)
leaf Setup -> Op + DetectShell + bash/zsh dumps or MergeInputs
root Setup  -> WorkDir under t.TempDir (no t.Setenv / t.Chdir)
root Run    -> lookpath.ResolveLoginEnvs | lookpath.MergeEnvs
               with injectable DetectShell + RunLogin
leaf Assert -> Shell/Envs or Merged + error + spies
```

## Preconditions

- Package under test:
  `github.com/xhd2015/dot-pkgs/go-pkgs/shell/lookpath`
- **Classic TDD:** public symbols under test do **not** exist yet; root `Run`
  imports them so the suite is **RED** (compile fail) until implementer lands:
  - `ResolveLoginEnvs`
  - `MergeEnvs`
  - `LoginEnvOptions.DetectShell`
- Existing `Resolve*LoginEnv(s)` and `./tests/login-env/` stay sealed GREEN.
- All default leaves are L2: injectable `RunLogin` + `DetectShell` only. No
  real login shell, no process PATH / env mutation.
- Parallel-safe isolation: `t.TempDir()` only; no `os.Setenv` / `t.Setenv` /
  `os.Chdir` / `t.Chdir` for config.

## Steps

1. Root `Setup` allocates `WorkDir` and zeros inject fixtures.
2. Grouping `Setup` sets `Op` and detection branch context.
3. Leaf `Setup` fills dump fixtures / MergeInputs / fail flags.
4. Root `Run` dispatches to `ResolveLoginEnvs` or `MergeEnvs`.
5. Leaf `Assert` checks Shell/Envs/Merged/error and spies.

## Context

- Production dump: `env -0` (NUL-delimited `KEY=value`); harness injects the
  same wire format via `BashStdout` / `ZshStdout` and `nulEnvDump`.
- Detect is always injected (never read process `$SHELL` in this suite).
- Product must not mutate process env/cwd.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.WorkDir = t.TempDir()
	req.Op = ""
	req.Home = filepath.Join(req.WorkDir, "home")
	req.Timeout = 0
	req.ShellBin = ""
	req.DetectShellResult = ""
	req.BashStdout = ""
	req.BashFail = false
	req.ZshStdout = ""
	req.ZshFail = false
	req.MergeInputs = nil
	req.MergeNoArgs = false
	return nil
}
```
