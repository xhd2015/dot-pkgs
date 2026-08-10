# Scenario

**Feature**: path-bound iTerm2 tell header + injectable ResolveAppPath

```
# resolve
ITERM2_APP_PATH / ~/Applications / /Applications
  -> ResolveAppPathWith(Getenv, Home, IsApp) -> app path or ""

# tell header
appPath -> TellApplicationHeader -> path-bound quoted path | bare "iTerm2"

# scripts
appPath + dir -> Build*App -> AppleScript starting with same header
```

## Preconditions

- Package gains Classic TDD surface locked in `DOCTEST.md`:
  - `EnvITerm2AppPath` (optional export; name `"ITERM2_APP_PATH"`)
  - `ResolveAppPathOpts` + `ResolveAppPathWith`
  - `TellApplicationHeader`
  - `BuildForceNewWindowScriptApp` / `BuildScriptApp` / `BuildPathScanSmokeScriptApp`
- Parallel-safe: inject Getenv/Home/IsApp via Request → Run; **no** `t.Setenv`.
- No live iTerm / dual-install E2E in this tree.

## Steps

1. Leaves set `req.Phase` and fixture fields.
2. Root Run calls product APIs; Assert checks paths / script substrings.

## Context

- Nested doctest root — independent of parent open-dir, focus, and tab-set trees.
- Env-missing matches localbot: set but unusable → empty, **no** fallthrough.
- Path-bound shape (string literal so AppleScript loads iTerm's dictionary):
  `tell application "<escaped path>"`
- Do **not** use `POSIX file "…" as text` as the tell target — that is a runtime
  expression and fails to compile iTerm terms (`create window with default profile`).

```go
import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

// systemApp is the locked system candidate path.
const systemApp = "/Applications/iTerm.app"

func homeApp(home string) string {
	return filepath.Join(home, "Applications", "iTerm.app")
}

// bareTellTarget is the fallback application name when resolve is empty.
func bareTellTarget() string {
	return `tell application "iTerm2"`
}

// pathBoundTellLine is the expected path-bound tell line for appPath.
func pathBoundTellLine(appPath string) string {
	esc := iterm2.EscapePathForAppleScript(appPath)
	return `tell application "` + esc + `"`
}

// hasPathBoundTell reports whether s uses quoted-path tell for appPath.
func hasPathBoundTell(s, appPath string) bool {
	if appPath == "" {
		return false
	}
	if strings.Contains(s, pathBoundTellLine(appPath)) {
		return true
	}
	// Accept EscapePathForAppleScript embedding even if whitespace differs slightly.
	esc := iterm2.EscapePathForAppleScript(appPath)
	return strings.Contains(s, `tell application "`+esc+`"`)
}

// hasBareTellTarget reports bare tell application "iTerm2".
func hasBareTellTarget(s string) bool {
	return strings.Contains(s, bareTellTarget())
}

// assertPathBoundScript requires path-bound tell and rejects bare iTerm2 target.
func assertPathBoundScript(t *testing.T, script, appPath string) {
	t.Helper()
	if script == "" {
		t.Fatal("expected non-empty script")
	}
	if !hasPathBoundTell(script, appPath) {
		t.Fatalf("script must path-bound tell for %q; want substring like %q; script:\n%s",
			appPath, pathBoundTellLine(appPath), script)
	}
	if hasBareTellTarget(script) {
		t.Fatalf("path-bound script must not use bare %q; script:\n%s", bareTellTarget(), script)
	}
	if strings.Contains(script, "POSIX file") {
		t.Fatalf("path-bound script must not use POSIX file expression; script:\n%s", script)
	}
}
```
