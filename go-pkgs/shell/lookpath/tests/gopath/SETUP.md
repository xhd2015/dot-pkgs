# Scenario

**Feature**: resolve GOPATH via login shells, go env, then ~/go fallback

```
# L2 harness (parallel-safe)
leaf Setup -> bash/zsh dumps + LookPath/RunGoEnv/Home inject
root Setup  -> WorkDir + Home under t.TempDir (no t.Setenv / t.Chdir)
root Run    -> lookpath.ResolveGoPathWith(GoPathOptions)
               LoginEnv.RunLogin / LookPath / RunGoEnv injected
leaf Assert -> Path + soft/hard error + cascade spies
```

## Preconditions

- Package under test:
  `github.com/xhd2015/dot-pkgs/go-pkgs/shell/lookpath`
- **Classic TDD:** public symbols under test do **not** exist yet; root `Run`
  imports them so the suite is **RED** (compile fail) until implementer lands:
  - `GoPathOptions`
  - `ResolveGoPath`, `ResolveGoPathWith`
- P1 login-env APIs (`LoginEnvOptions`, `ResolveBashLoginEnv`,
  `ResolveZshLoginEnv`) already exist and are GREEN; product cascade may call
  them with `opts.LoginEnv` (tests inject `RunLogin` only).
- All default leaves are L2: injectable `RunLogin` / `LookPath` / `RunGoEnv` /
  `Home` only. No real login shell, no real `go env`, no process PATH / env
  mutation.
- Parallel-safe isolation: `t.TempDir()` only; no `os.Setenv` / `t.Setenv` /
  `os.Chdir` / `t.Chdir` for config.
- Sibling suites `./tests/lookpath/`, `./tests/lookup-paths/`,
  `./tests/login-env/` are out of scope (do not modify).

## Steps

1. Root `Setup` allocates `WorkDir` and default `Home` under the temp root;
   zeros inject fixtures.
2. Grouping `Setup` documents the cascade winner branch.
3. Leaf `Setup` fills bash/zsh dumps, LookPath, RunGoEnv, and/or soft-fail flags.
4. Root `Run` calls `ResolveGoPathWith` with full injectors and records spies.
5. Leaf `Assert` checks Path / error and cascade short-circuit spies.

## Context

- Login dumps use production-style `env -0` (NUL-delimited `KEY=value`) via
  `nulEnvDump`.
- Multi-GOPATH: first colon-separated segment after TrimSpace, then
  `filepath.Clean`.
- Soft failures never abort the cascade before home fallback.
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
	req.Home = filepath.Join(req.WorkDir, "home")
	req.BashStdout = ""
	req.BashFail = false
	req.ZshStdout = ""
	req.ZshFail = false
	req.GoBin = ""
	req.LookPathFail = false
	req.GoEnvStdout = ""
	req.GoEnvFail = false
	return nil
}
```
