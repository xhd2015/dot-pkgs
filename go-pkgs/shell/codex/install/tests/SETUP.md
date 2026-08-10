# Scenario

**Feature**: OpenAI Codex CLI install library (parse → latest → local → install/update → ensure)

```
# L2 harness (parallel-safe)
leaf Setup -> Operation + version/HTTP/ensure fixtures on Request
root Setup  -> WorkDir under t.TempDir (no t.Setenv / t.Chdir)
root Run    -> install.* public API with injectable HTTPClient / LookPath /
               RunShell / RunVersion / FetchLatest
leaf Assert -> Response fields + shell/fetch call spies
```

## Preconditions

- Package under test:
  `github.com/xhd2015/dot-pkgs/go-pkgs/shell/codex/install`
- **Classic TDD:** package and public symbols do **not** exist yet; root `Run`
  imports them so the suite is **RED** (compile fail) until implementer lands:
  - `ParseVersion`, `NeedsUpdate`, `LatestVersion`, `LocalVersion`
  - `Install`, `Update`, `Ensure`
  - `InstallScriptURL`, `InstallCmd`, `UpdateCmd`, `NPMLatestURL`
  - opts types: `LatestVersionOpts`, `LocalVersionOpts`, `InstallOpts`,
    `UpdateOpts`, `EnsureOpts`, `Result`
- All default leaves are L2: injectable HTTP (`httptest`), LookPath, RunShell,
  RunVersion, FetchLatest. No real network, no real `codex` binary.
- Parallel-safe isolation: `t.TempDir()` only; no `os.Setenv` / `t.Setenv` /
  `os.Chdir` / `t.Chdir` for config.
- Out of suite: real npm, GitHub releases fallback, spl landing wiring.

## Steps

1. Root `Setup` allocates `WorkDir`.
2. Grouping `Setup` sets `Operation`.
3. Leaf `Setup` sets scenario fields (version strings, HTTPMode, Ensure*).
4. Root `Run` dispatches to the public API and returns `Response`.
5. Leaf `Assert` checks error/success and spy call counts/arguments.

## Context

- Install recipe constant matches landing dry-run:
  `curl -fsSL https://chatgpt.com/codex/install.sh | sh`
- NPM latest endpoint:
  `https://registry.npmjs.org/@openai/codex/latest` → JSON `version`
- Ensure must not call latest fetch when bin is missing.
- Unknown latest/local → noop (no forced error).

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
	req.HTTPMode = ""
	req.NPMVersion = "0.147.0"
	return nil
}
```
