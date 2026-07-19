# Scenario

**Feature**: tab-set library — create (P1), find+busy (P2), run/status/stop (P3)

```
# P1 create
caller TabSetSpec -> BuildTabSetNewWindowScript -> AppleScript

# P2 find / busy
caller setName -> BuildTabSetFindScript / ParseTabSetFindOutput
caller fgComm -> ClassifyBusyFromComm

# P3 orchestration (injectable TabSetConfig)
caller spec + opts + cfg{Find,Busy,Exec} -> RunTabSet / StatusTabSet / StopTabSet
```

## Preconditions

- Package `github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2` is importable.
- **P1+P2 landed:** create/find/busy APIs as used by existing leaves.
- **P3 Classic TDD:** `RunTabSet`, `StatusTabSet`, `StopTabSet`, `TabSetConfig`,
  `TabSetRunMode`, result types, `ErrNoITermWindow`.

## Steps

1. Default `req.Phase` to `build-tab-set-script` when empty (P1 leaves).
2. P2/P3 grouping or leaf Setups override Phase.

## Context

- P1 asserts use `resp.Script` from Run.
- P2/P3 asserts call product APIs directly (package isolation).
- No live iTerm for CI leaves — inject Find/Busy/Exec.

```go
import (
	"strings"
	"testing"
)

// Session variable name literals stamped / scanned by tab-set scripts.
const (
	tabSetVarLiteral    = `user.koolTabSet`
	tabSetTabVarLiteral = `user.koolTabSetTab`
)

func Setup(t *testing.T, req *Request) error {
	if req.Phase == "" {
		req.Phase = "build-tab-set-script"
	}
	return nil
}

func countOccurrences(s, sub string) int {
	if sub == "" {
		return 0
	}
	return strings.Count(s, sub)
}

func writeTextLine(command string) string {
	return `write text "` + command + `"`
}

func joinedScripts(scripts []string) string {
	return strings.Join(scripts, "\n---\n")
}

func scriptsContainCtrlC(scripts []string) bool {
	for _, s := range scripts {
		lower := strings.ToLower(s)
		if strings.Contains(lower, "control-c") ||
			strings.Contains(lower, "ctrl-c") ||
			strings.Contains(lower, `keystroke "c" using control`) ||
			strings.Contains(lower, "keystroke \"c\" using control") ||
			strings.Contains(lower, "key code 8 using control") {
			return true
		}
	}
	return false
}
```
