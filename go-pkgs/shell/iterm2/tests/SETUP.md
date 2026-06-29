# Scenario

**Feature**: `iterm2` builds AppleScript and opens directories in iTerm2 on macOS

```
# open pipeline
caller dir + follow-ups -> OpenConfig -> validate -> BuildScript -> osascript -> iTerm2

# script-only (tests)
caller dir + follow-ups -> BuildScript -> AppleScript string
```

## Preconditions

- Package `github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2` is importable.
- Implementer adds test hooks: `SetGOOSForTest`, `FollowUpCommands` on `Config`,
  `EscapeCommandForAppleScript`, smart `BuildScript`, `BuildPathScanSmokeScript`,
  and path normalization in `OpenConfig`.

## Context

- Script assertions use substring checks aligned with vscode-ext `iterm2.ts`.
- `open/` leaves use injectable `Config` — no real iTerm2 side effects.
- Live leaf requires macOS, iTerm2 installed, and `--label side-effect-open-iterm2`.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func absDir(t *testing.T, dir string) string {
	t.Helper()
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func scriptHasPathScan(script string) bool {
	return strings.Contains(script, `variable named "path"`) &&
		strings.Contains(script, "matchingWindow")
}

func scriptUsesTellSession(script string) bool {
	return strings.Contains(script, "tell aSession") &&
		!strings.Contains(script, `variable named "path" of aSession`)
}

func scriptUsesOnError(script string) bool {
	return strings.Contains(script, "on error")
}
```