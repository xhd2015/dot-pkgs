# shell/iterm2 — Smart Open iTerm2 Library

## Version
0.0.2

Doc tests for `github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2`. The library
builds AppleScript to activate iTerm2, reuse a window via session `path` scan
when possible, `cd` to a target directory, optionally run follow-up commands,
and run `osascript`. Most leaves inject dependencies; one live leaf exercises
real `osascript` when labeled.

P1 **tab-set** create scripts live in the nested root `tab-set/` (own
`DOCTEST.md` / `Request` / `Run`) so Classic TDD missing-API symbols do not
break this tree’s existing leaves.

P1 **app-path** (path-bound tell + `ITERM2_APP_PATH` resolve) lives in nested
root `app-path/` (own `DOCTEST.md` / `Request` / `Run`) for the same reason.

## DSN (Domain Specific Notion)

### Participants

- **Caller** — supplies a directory path, optional follow-up shell commands,
  and optional `Config` overrides for tests.
- **Path normalizer** — resolves absolute paths and `filepath.EvalSymlinks`
  when the path exists (canonical target for smart-open matching).
- **Script builder** — `BuildScript` emits AppleScript: scan windows/tabs/sessions,
  read session variable `path` via `tell aSession`, reuse matching window with
  `create tab`, else `create window`; writes `cd` and follow-up `write text` lines.
- **Tab-set script builder** — see nested root `tab-set/` (`BuildTabSetNewWindowScript`).
- **Escaper** — `EscapePathForAppleScript` and `EscapeCommandForAppleScript`
  escape backslashes and double quotes for embedded string literals.
- **OpenConfig** — validates platform (darwin), install check, directory stat,
  builds script, invokes `Config.Osascript` or default `osascript`.
- **osascript runner** — macOS `/usr/bin/osascript -e` (or injectable fake).

### Behaviors

**BuildScript**

- Always `activate` iTerm2 and set `targetDir` to escaped absolute path string.
- Scan non-hotkey windows; compare session `path` to `targetDir`.
- On match → `create tab with default profile` in that window; else new window.
- Session commands: `write text ("cd " & quoted form of targetDir)` then each
  follow-up as `write text "<escaped command>"`.
- Must not emit `exec $SHELL`.
- Path probe uses `tell aSession` with `on error` handler (not invalid `of aSession`).

**OpenConfig**

- Non-darwin → `ErrUnsupportedPlatform`.
- Not installed → `ErrNotInstalled`.
- Missing path or not a directory → wrapped stat / not-a-directory errors.
- Success → non-empty script passed to osascript runner.

**Live smoke**

- `BuildPathScanSmokeScript` probes session paths only; returns `"ok"` on success.

**Tab-set (nested)**

- Nested doctest root `tab-set/` covers `BuildTabSetNewWindowScript` (P1 Classic TDD).

**Contents (nested)**

- Nested doctest root `contents/` covers `BuildContentsScript` / `Contents`
  (no-focus dump; home then system; skip not-running).

**App-path (nested)**

- Nested doctest root `app-path/` covers injectable `ResolveAppPathWith`
  (`ITERM2_APP_PATH` → home → system; env-missing no fallthrough),
  `TellApplicationHeader` (path-bound vs bare), and representative Build\*App
  scripts. Classic TDD **RED** until implementer lands path-bound + env resolve.

## Decision Tree

```
iterm2-lib/
├── script/                         [Phase=build-script]
│   ├── smart-open-branches/        path scan + tab reuse + window fallback
│   ├── follow-up-single/           one follow-up after cd
│   ├── follow-up-multiple/         ordered follow-ups
│   ├── uses-tell-session/          tell aSession path access
│   ├── reuse-current-session/      -r: path scan; focus match, window+cd miss
│   ├── reuse-registers-user-variable/  miss branch sets user.koolTargetDir
│   ├── reuse-scans-user-variable/      scan matches path or user.koolTargetDir
│   ├── smart-open-match-cd-scoped/     match branch cd scoped to matchingWindow tab
│   ├── smart-open-scans-user-variable/ scan matches path or user.koolTargetDir
│   ├── reuse-match-selects-window/     reuse match selects matchingWindow to front
│   └── no-exec-shell/              no exec $SHELL
├── tab-set/                        [nested DOCTEST.md — tab-set create/find/run]
├── app-path/                       [nested DOCTEST.md — resolve + path-bound tell, Classic TDD RED]
│   ├── resolve/                    env-wins, prefer-home, system-only, empty, env-missing
│   ├── tell-header/                path-bound, bare-fallback
│   └── script/                     force-new / smart-open / smoke path-bound
├── escaping/                       [Phase=escape-*]
│   ├── path-quotes/                EscapePathForAppleScript
│   └── command-quotes/             EscapeCommandForAppleScript
├── open/                           [Phase=open-config, mocked runner]
│   ├── invokes-osascript/          captures non-empty script
│   ├── not-directory/              file path error
│   ├── not-installed/              ErrNotInstalled
│   ├── nonexistent-path/           stat error
│   └── osascript-failure/          wrapped osascript error
├── normalize/                      [Phase=open-config, symlink dir]
│   └── symlink-resolves/           EvalSymlinks in script target
└── live/                           [Phase=live-smoke]
    └── scan-smoke/                 real osascript (labeled)
```

## Test Index

| Leaf | Phase | Description |
|------|-------|-------------|
| `script/smart-open-branches/` | build-script | Scan, create tab, create window fallback |
| `script/follow-up-single/` | build-script | Single follow-up write text |
| `script/follow-up-multiple/` | build-script | Multiple follow-ups in order |
| `script/uses-tell-session/` | build-script | `tell aSession` + `on error` |
| `script/reuse-current-session/` | build-script | Reuse: scan, focus match, window fallback |
| `script/reuse-registers-user-variable/` | build-script | Reuse miss branch sets `user.koolTargetDir` |
| `script/reuse-scans-user-variable/` | build-script | Reuse scan matches `path` or `user.koolTargetDir` |
| `script/smart-open-match-cd-scoped/` | build-script | Smart match cd scoped to matchingWindow tab |
| `script/smart-open-scans-user-variable/` | build-script | Smart scan matches `path` or `user.koolTargetDir` |
| `script/reuse-match-selects-window/` | build-script | Reuse match selects matchingWindow to front |
| `script/no-exec-shell/` | build-script | No `exec $SHELL` |
| `tab-set/*` (nested root) | build-tab-set-script | See `tab-set/DOCTEST.md` (P1 Classic TDD) |
| `app-path/*` (nested root) | resolve-app / tell-header / build-*-app | See `app-path/DOCTEST.md` (path-bound + env resolve, Classic TDD RED) |
| `escaping/path-quotes/` | escape-path | Escapes `"` in paths |
| `escaping/command-quotes/` | escape-command | Escapes `"` in commands |
| `open/invokes-osascript/` | open-config | Injectable osascript receives script |
| `open/not-directory/` | open-config | File → not a directory |
| `open/not-installed/` | open-config | Installed=false → ErrNotInstalled |
| `open/nonexistent-path/` | open-config | Missing path stat error |
| `open/osascript-failure/` | open-config | Osascript error propagated |
| `normalize/symlink-resolves/` | open-config | Symlink resolved in captured script |
| `live/scan-smoke/` | live-smoke | Real path scan smoke (label) |

## How to Run

```sh
doctest vet ./dot-pkgs-kool/go-pkgs/shell/iterm2/tests
doctest test ./dot-pkgs-kool/go-pkgs/shell/iterm2/tests
doctest test --label side-effect-open-iterm2 ./dot-pkgs-kool/go-pkgs/shell/iterm2/tests
```

```go
import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

type Request struct {
	Phase                string
	Dir                  string
	Mode                 string // "reuse" for BuildReuseCurrentSessionScript
	FollowUps            []string
	EscapeInput          string
	UseInstalledOverride bool
	InstalledOK          bool
	OsascriptFail        bool
	ForceGOOS            string
}

type Response struct {
	Script         string
	Escaped        string
	CapturedScript string
	LiveStdout     string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	resp := &Response{}
	switch req.Phase {
	case "build-script":
		if req.Mode == "reuse" {
			resp.Script = iterm2.BuildReuseCurrentSessionScript(req.Dir, req.FollowUps...)
		} else {
			resp.Script = iterm2.BuildScript(req.Dir, req.FollowUps...)
		}
		return resp, nil
	case "escape-path":
		resp.Escaped = iterm2.EscapePathForAppleScript(req.EscapeInput)
		return resp, nil
	case "escape-command":
		resp.Escaped = iterm2.EscapeCommandForAppleScript(req.EscapeInput)
		return resp, nil
	case "open-config":
		goos := req.ForceGOOS
		if goos == "" && runtime.GOOS != "darwin" {
			goos = "darwin"
		}
		if goos != "" {
			iterm2.SetGOOSForTest(goos)
			t.Cleanup(func() { iterm2.SetGOOSForTest("") })
		}
		var captured string
		cfg := &iterm2.Config{
			FollowUpCommands: req.FollowUps,
			Installed: func() bool {
				if req.UseInstalledOverride {
					return req.InstalledOK
				}
				return true
			},
			Osascript: func(script string) error {
				captured = script
				if req.OsascriptFail {
					return errors.New("mock osascript failed")
				}
				return nil
			},
		}
		err := iterm2.OpenConfig(req.Dir, cfg)
		resp.CapturedScript = captured
		return resp, err
	case "live-smoke":
		if runtime.GOOS != "darwin" {
			t.Skip("live smoke requires darwin")
		}
		if !iterm2.IsInstalled() {
			t.Skip("iTerm2 not installed")
		}
		script := iterm2.BuildPathScanSmokeScript()
		cmd := exec.Command("osascript", "-e", script)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return resp, fmt.Errorf("osascript: %w\n%s", err, out)
		}
		resp.LiveStdout = strings.TrimSpace(string(out))
		resp.Script = script
		return resp, nil
	default:
		return nil, fmt.Errorf("unknown phase %q", req.Phase)
	}
}
```