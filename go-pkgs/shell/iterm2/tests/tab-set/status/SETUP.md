# Scenario

**Feature**: `StatusTabSet` reports per-tab state from Find + Busy

```
spec.Tabs + Find + Busy map -> StatusTabSet -> TabSetStatus (running|idle|missing|unknown)
```

## Preconditions

- Product exports `StatusTabSet`, `TabSetStatus`, `TabStatusEntry`.

## Steps

1. Set Phase `status-tab-set`.
2. Leaves configure Tabs, FindSessions, BusyByTab.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Setup(t *testing.T, req *Request) error {
	req.Phase = "status-tab-set"
	if req.TabSetName == "" {
		req.TabSetName = "bots"
	}
	return nil
}

func buildStatusConfig(req *Request) *iterm2.TabSetConfig {
	refs := make([]iterm2.TabSessionRef, len(req.FindSessions))
	for i, s := range req.FindSessions {
		refs[i] = iterm2.TabSessionRef{
			SetName:   s.SetName,
			TabID:     s.TabID,
			WindowID:  s.WindowID,
			SessionID: s.SessionID,
			TTY:       s.TTY,
		}
		if refs[i].SetName == "" {
			refs[i].SetName = req.TabSetName
		}
	}
	busyMap := req.BusyByTab
	if busyMap == nil {
		busyMap = map[string]string{}
	}
	return &iterm2.TabSetConfig{
		Find: func(setName string) ([]iterm2.TabSessionRef, error) {
			return refs, nil
		},
		Busy: func(ref iterm2.TabSessionRef) iterm2.BusyState {
			switch strings.ToLower(strings.TrimSpace(busyMap[ref.TabID])) {
			case "idle":
				return iterm2.BusyStateIdle
			case "busy":
				return iterm2.BusyStateBusy
			default:
				return iterm2.BusyStateUnknown
			}
		},
		Exec: func(script string) error { return nil },
	}
}

func statusStateFor(st *iterm2.TabSetStatus, tabID string) string {
	if st == nil {
		return ""
	}
	for _, e := range st.Tabs {
		if e.TabID == tabID {
			return e.State
		}
	}
	return ""
}
```
