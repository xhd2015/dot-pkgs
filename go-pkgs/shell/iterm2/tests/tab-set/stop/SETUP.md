# Scenario

**Feature**: `StopTabSet` closes marked sessions or warns when not running

```
setName + TabSetConfig{Find, Exec}
  -> StopTabSet
  -> TabSetStopResult (ClosedWindows/Tabs, Warning)
```

## Preconditions

- Product exports `StopTabSet`, `TabSetStopResult`.

## Steps

1. Set Phase `stop-tab-set`.
2. Leaves configure FindSessions (empty vs marked).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {

	req.Phase = "stop-tab-set"
	if req.TabSetName == "" {
		req.TabSetName = "bots"
	}
	return nil
}

func buildStopConfig(req *Request, scripts *[]string) *iterm2.TabSetConfig {
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
	return &iterm2.TabSetConfig{
		Find: func(setName string) ([]iterm2.TabSessionRef, error) {
			return refs, nil
		},
		Busy: func(ref iterm2.TabSessionRef) iterm2.BusyState {
			return iterm2.BusyStateIdle
		},
		Exec: func(script string) error {
			if scripts != nil {
				*scripts = append(*scripts, script)
			}
			return nil
		},
	}
}
```
