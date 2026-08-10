# Scenario

**Feature**: iTerm2 official stable install library (resolve → download → extract → install → verify)

```
# L2 harness (parallel-safe)
leaf Setup -> Operation + HTTPMode / fixtures on Request
root Setup  -> WorkDir + Home under t.TempDir (no t.Setenv / t.Chdir)
root Run    -> install.* public API with injectable HTTPClient / Home / Runner /
               Register / Open / ClearQuarantineFn
leaf Assert -> Response fields + filesystem under WorkDir/Home
```

## Preconditions

- Package under test:
  `github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/install`
- All default leaves are L2: injectable HTTP (`httptest`), temp Home/Target, fake
  zip/app fixtures. No real network, no real `osascript` / `lsregister` /
  `open` / `xattr`.
- Parallel-safe isolation: `t.TempDir()` only; no `os.Setenv` / `t.Setenv` /
  `os.Chdir` / `t.Chdir` for config.
- Platform: filesystem ops work on macOS + Linux; full product scriptable path is
  darwin-oriented but skipped via injection in this suite.
- InstallViaUserOpen leaves stay **RED** until product opts/hooks exist.

## Steps

1. Root `Setup` allocates `WorkDir` and default `Home` (`WorkDir/home`).
2. Grouping `Setup` sets `Operation`.
3. Leaf `Setup` sets scenario fields (`HTTPMode`, targets, fixture flags).
4. Root `Run` dispatches to the public API and returns `Response`.
5. Leaf `Assert` checks error/success and observable filesystem/HTTP outcomes.

## Context

- Mini-bundle fixture: `iTerm.app/Contents/Info.plist` + `Contents/MacOS/iTerm2`.
- Zip fixtures built in-process via `archive/zip` (no binary fixtures in git).
- Default install target preference: `~/Applications/iTerm.app` via injected Home.
- Universal zip: resolve asserts URL has no `arm64` / `amd64` substrings.
- `InstallOpts.LatestURL` / `ResolveOpts.LatestURL` inject fake latest endpoint.

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
	req.Operation = ""
	req.HTTPMode = ""
	req.FinalZipName = "iTerm2-3_6_11.zip"
	req.SkipScriptable = true
	return nil
}
```
