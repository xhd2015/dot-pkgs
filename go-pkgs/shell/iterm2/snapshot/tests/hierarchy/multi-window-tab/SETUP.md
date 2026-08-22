# Scenario

**Feature**: multi-window multi-tab hierarchy indices and names preserved

```
W1/T1/S + W2/T1/S + W2/T2/S -> Capture -> structure and indices match fixture
```

## Steps

1. Two windows: window 1 with one tab/session; window 2 with two tabs/sessions.
2. All sessions idle so enrich does not obscure hierarchy asserts.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Windows = []snapshot.SnapshotWindow{
		{
			Index: 1, Name: "Win-A", WindowID: 10,
			Tabs: []snapshot.SnapshotTab{
				{
					Index: 1, Name: "A1",
					Sessions: []snapshot.SnapshotSession{
						baseSession(1, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "s-a1", "/dev/ttys010", "Default"),
					},
				},
			},
		},
		{
			Index: 2, Name: "Win-B", WindowID: 20,
			Tabs: []snapshot.SnapshotTab{
				{
					Index: 1, Name: "B1",
					Sessions: []snapshot.SnapshotSession{
						baseSession(1, "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", "s-b1", "/dev/ttys020", "Default"),
					},
				},
				{
					Index: 2, Name: "B2",
					Sessions: []snapshot.SnapshotSession{
						baseSession(1, "cccccccc-cccc-cccc-cccc-cccccccccccc", "s-b2", "/dev/ttys021", "Default"),
					},
				},
			},
		},
	}
	req.IdleTTYs = []string{"ttys010", "ttys020", "ttys021"}
	return nil
}
```
