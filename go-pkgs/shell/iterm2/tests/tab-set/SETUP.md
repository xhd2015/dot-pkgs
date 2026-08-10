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
- **P1+P2+P3 landed:** create/find/busy/orchestration APIs as used by existing leaves.
- **NoSubmit Classic TDD:** `TabSpec.NoSubmit` + `write text "…" without newline`
  on command paths (new-window, create-tab, resend); new leaves RED until implementer.

## Steps

1. Default `req.Phase` to `build-tab-set-script` when empty (P1 leaves).
2. P2/P3 grouping or leaf Setups override Phase.

## Context

- P1 asserts use `resp.Script` from Run.
- P2/P3 asserts call product APIs directly (package isolation).
- No live iTerm for CI leaves — inject Find/Busy/Exec.

```go
import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

// Session variable name literals stamped / scanned by tab-set scripts.
const (
	tabSetVarLiteral    = `user.koolTabSet`
	tabSetTabVarLiteral = `user.koolTabSetTab`
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
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

// writeTextWithoutNewline is the AppleScript form when TabSpec.NoSubmit is true.
func writeTextWithoutNewline(command string) string {
	return `write text "` + command + `" without newline`
}

// commandWriteHasWithoutNewline reports whether script contains a write-text of
// command with the without-newline qualifier (NoSubmit path).
func commandWriteHasWithoutNewline(script, command string) bool {
	return strings.Contains(script, writeTextWithoutNewline(command))
}

// commandWriteSubmits reports a submit write (with newline/Enter) for command:
// a write text line for the command that is not the without-newline form.
func commandWriteSubmits(script, command string) bool {
	if commandWriteHasWithoutNewline(script, command) {
		return false
	}
	return strings.Contains(script, writeTextLine(command)) ||
		strings.Contains(script, `write text "`+command)
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

// setNoSubmit sets TabSpec.NoSubmit via reflection so NoSubmit leaves compile
// before the product field exists (Classic TDD: assert RED, not build fail).
// Returns an error if the field is missing or not settable.
func setNoSubmit(tab *iterm2.TabSpec, v bool) error {
	if tab == nil {
		return fmt.Errorf("nil TabSpec")
	}
	rv := reflect.ValueOf(tab).Elem()
	f := rv.FieldByName("NoSubmit")
	if !f.IsValid() {
		return fmt.Errorf("TabSpec.NoSubmit missing")
	}
	if !f.CanSet() {
		return fmt.Errorf("TabSpec.NoSubmit not settable")
	}
	if f.Kind() != reflect.Bool {
		return fmt.Errorf("TabSpec.NoSubmit kind %s, want bool", f.Kind())
	}
	f.SetBool(v)
	return nil
}

// mustSetNoSubmit is like setNoSubmit but fails the test on error.
func mustSetNoSubmit(t *testing.T, tab *iterm2.TabSpec, v bool) {
	t.Helper()
	if err := setNoSubmit(tab, v); err != nil {
		t.Fatal(err)
	}
}
```