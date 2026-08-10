# Scenario

**Feature**: pure iTerm session list, find-by-TTY, and focus (P1)

```
# list + parse
BuildSessionListScript -> osascript dump -> ParseSessionListOutput -> []SessionRef

# find
[]SessionRef + query TTYs -> FindByTTY (via NormalizeTTY) -> matching refs

# focus
SessionRef -> BuildFocusScript / Focus(cfg.Exec) -> activate + select window + tab
```

## Preconditions

- Package `github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2` gains the P1 API
  locked in root `DOCTEST.md` (`SessionRef`, `NormalizeTTY`,
  `ParseSessionListOutput`, `BuildSessionListScript`, `FindByTTY`,
  `BuildFocusScript`, `FocusConfig`, `Focus`).
- Classic TDD: leaves RED until those symbols exist and behave as specified.
- No live iTerm required for GREEN (injectable Exec; pure parse/find).

## Steps

1. Default Phase left empty for grouping nodes; leaves set Phase.
2. Helpers convert `SessionRefInput` ↔ `iterm2.SessionRef` for Run/Assert.

## Context

- Nested doctest root — independent of open-dir and tab-set trees.
- Parallel-safe: no process env/cwd mutation; Focus uses mock Exec.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func sessionRefFromInput(in SessionRefInput) iterm2.SessionRef {
	return iterm2.SessionRef{
		WindowID:   in.WindowID,
		WindowName: in.WindowName,
		TabIndex:   in.TabIndex,
		SessionID:  in.SessionID,
		TTY:        in.TTY,
		Name:       in.Name,
	}
}

func sessionRefsFromInputs(in []SessionRefInput) []iterm2.SessionRef {
	out := make([]iterm2.SessionRef, len(in))
	for i, r := range in {
		out[i] = sessionRefFromInput(r)
	}
	return out
}

func sessionRefToInput(r iterm2.SessionRef) SessionRefInput {
	return SessionRefInput{
		WindowID:   r.WindowID,
		WindowName: r.WindowName,
		TabIndex:   r.TabIndex,
		SessionID:  r.SessionID,
		TTY:        r.TTY,
		Name:       r.Name,
	}
}

func sessionRefsToInputs(refs []iterm2.SessionRef) []SessionRefInput {
	if refs == nil {
		return nil
	}
	out := make([]SessionRefInput, len(refs))
	for i, r := range refs {
		out[i] = sessionRefToInput(r)
	}
	return out
}
```
