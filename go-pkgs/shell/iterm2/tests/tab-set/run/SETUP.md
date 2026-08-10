# Scenario

**Feature**: `RunTabSet` smart / new-window / no-new-window orchestration

```
spec + TabSetRunOptions + TabSetConfig{Find, Busy, Exec, FrontmostWindowID}
  -> RunTabSet
  -> TabSetRunResult (actions, warning, created window) + captured Exec scripts
```

## Preconditions

- Product exports `RunTabSet`, `TabSetConfig`, `TabSetRunOptions`, `TabSetRunMode`,
  `TabSetRunResult`, `TabRunResult`, `ErrNoITermWindow`.

## Steps

1. Set `req.Phase` to `run-tab-set`.
2. Leaves fill Tabs, FindSessions, BusyByTab, RunMode, FrontmostWindowID.

## Context

- All leaves inject Find/Busy/Exec — no live iTerm.
- Find order = recency (first WindowID wins when multi-window).

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {

	req.Phase = "run-tab-set"
	if req.RunMode == "" {
		req.RunMode = "smart"
	}
	if req.TabSetName == "" {
		req.TabSetName = "bots"
	}
	return nil
}

func tabSetSpecFromReq(req *Request) iterm2.TabSetSpec {
	tabs := make([]iterm2.TabSpec, len(req.Tabs))
	for i, tab := range req.Tabs {
		tabs[i] = iterm2.TabSpec{
			ID:      tab.ID,
			Name:    tab.Name,
			Command: tab.Command,
			Cwd:     tab.Cwd,
		}
	}
	return iterm2.TabSetSpec{
		Name:       req.TabSetName,
		WindowName: req.WindowName,
		Tabs:       tabs,
	}
}

func sessionRefsFromReq(req *Request) []iterm2.TabSessionRef {
	out := make([]iterm2.TabSessionRef, len(req.FindSessions))
	for i, s := range req.FindSessions {
		out[i] = iterm2.TabSessionRef{
			SetName:   s.SetName,
			TabID:     s.TabID,
			WindowID:  s.WindowID,
			SessionID: s.SessionID,
			TTY:       s.TTY,
		}
		if out[i].SetName == "" {
			out[i].SetName = req.TabSetName
		}
	}
	return out
}

func busyStateFromString(s string) iterm2.BusyState {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "idle":
		return iterm2.BusyStateIdle
	case "busy":
		return iterm2.BusyStateBusy
	default:
		return iterm2.BusyStateUnknown
	}
}

func runModeFromReq(req *Request) iterm2.TabSetRunMode {
	switch req.RunMode {
	case "new-window":
		return iterm2.TabSetRunNewWindow
	case "no-new-window":
		return iterm2.TabSetRunNoNewWindow
	default:
		return iterm2.TabSetRunSmart
	}
}

// buildRunConfig builds injectable TabSetConfig and captures Exec scripts.
func buildRunConfig(req *Request, scripts *[]string) *iterm2.TabSetConfig {
	refs := sessionRefsFromReq(req)
	busyMap := req.BusyByTab
	if busyMap == nil {
		busyMap = map[string]string{}
	}
	return &iterm2.TabSetConfig{
		Find: func(setName string) ([]iterm2.TabSessionRef, error) {
			return refs, nil
		},
		Busy: func(ref iterm2.TabSessionRef) iterm2.BusyState {
			if s, ok := busyMap[ref.TabID]; ok {
				return busyStateFromString(s)
			}
			return iterm2.BusyStateUnknown
		},
		Exec: func(script string) error {
			*scripts = append(*scripts, script)
			return nil
		},
		FrontmostWindowID: req.FrontmostWindowID,
	}
}

func actionForTab(result *iterm2.TabSetRunResult, tabID string) string {
	if result == nil {
		return ""
	}
	for _, tr := range result.Tabs {
		if tr.TabID == tabID {
			return tr.Action
		}
	}
	return ""
}
```
