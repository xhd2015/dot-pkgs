# Scenario

**Feature**: shell/iterm2 title scripts and API for session/window names

```
# script-only
caller sessionID + target + title -> BuildSetTitleScript / BuildGetTitleScript -> AppleScript

# API
caller title + target + ITERM_SESSION_ID -> SetTitle / GetTitle -> error or values
```

## Preconditions

- Package `github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2` gains:
  - `TitleTarget` (`TitleTargetSession`, `TitleTargetWindow`)
  - `BuildSetTitleScript(sessionID string, target TitleTarget, title string) string`
  - `BuildGetTitleScript(sessionID string, target TitleTarget) string`
  - `SetTitle(title string, target TitleTarget) (old, new string, err error)`
  - `GetTitle(target TitleTarget) (string, error)`
- Session identity uses UUID after `:` in `ITERM_SESSION_ID`.
- Escaping reuses `EscapePathForAppleScript` (or equivalent) for titles.

## Context

- Nested doctest root — independent of open-dir library tests.
- API leaves that need no real osascript only cover env/validation errors.
- Script leaves assert UUID + session vs window targeting substrings.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.SessionID == "" {
		req.SessionID = defaultTitleSessionID
	}
	if req.Target == "" {
		req.Target = "session"
	}
	return nil
}

func scriptHasUUID(script, sid string) bool {
	return strings.Contains(script, sessionUUIDFromID(sid))
}
```
