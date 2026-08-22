# Scenario

**Feature**: injectable iTerm2 snapshot capture with process enrich (no agent)

```
# gate
Caller -> Collector.ITermRunning -> error when iTerm absent

# happy path
Caller -> NewCollector + ApplyPhasedFixture -> Capture
  -> windows/tabs/sessions + idle/busy/cwd enrich -> Snapshot + warnings
```

## Preconditions

- Package under test:
  `github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot` (absent until
  implementer — Classic TDD RED).
- All leaves are **L2 in-process**: inject via per-test `Collector` +
  `ApplyPhasedFixture` / field overrides. **No** real `osascript`, `ps`, or
  `lsof`. **No** process-global collector mutation.
- Parallel-safe: each leaf builds its own `Collector` in root `Run`.
- Agent fields absent from model and API.

## Steps

1. Root `Setup` zeros request fields for the leaf chain.
2. Grouping/leaf `Setup` sets hierarchy, idle/busy ttys, cwd, or inject modes.
3. Root `Run` builds `NewCollector()`, applies phased fixture, optional
   ListProcs override, then `Capture` or `CaptureWith`.
4. Leaf `Assert` checks snapshot / warnings / error only.

## Context

- Locked API: `Collector`, `NewCollector`, `Capture`, `CaptureWith`,
  `ApplyPhasedFixture`, `PhasedFixtureOpts`, model types without Agent.
- Error gate: iTerm not running → message includes `iTerm2 is not running`.
- Soft path: ListProcs/ListCwds failures and empty proc lists → warnings,
  Capture still returns a snapshot when hierarchy is available.

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ITermRunning = false
	req.Windows = nil
	req.IdleTTYs = nil
	req.BusyTTYs = nil
	req.BusyLeafByTTY = nil
	req.CwdByTTY = nil
	req.Now = time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	req.Hostname = "testhost"
	req.UseCaptureWith = false
	req.SpaceAllow = nil
	req.AppTag = ""
	req.ListProcsMode = ""
	req.ListProcsErr = ""
	return nil
}

// oneSessionWindow builds a single-window / single-tab / single-session fixture.
func oneSessionWindow(winIndex int, winName string, windowID uint64, tabIndex int, tabName string, sess snapshot.SnapshotSession) snapshot.SnapshotWindow {
	return snapshot.SnapshotWindow{
		Index:    winIndex,
		Name:     winName,
		WindowID: windowID,
		Tabs: []snapshot.SnapshotTab{
			{
				Index:    tabIndex,
				Name:     tabName,
				Sessions: []snapshot.SnapshotSession{sess},
			},
		},
	}
}

func baseSession(index int, id, name, tty, profile string) snapshot.SnapshotSession {
	return snapshot.SnapshotSession{
		Index:   index,
		ID:      id,
		Name:    name,
		TTY:     tty,
		Profile: profile,
	}
}
```
