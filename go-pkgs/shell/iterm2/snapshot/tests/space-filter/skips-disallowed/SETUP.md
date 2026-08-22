# Scenario

**Feature**: SpaceAllow skips non-matching FixedSpace window entirely

```
only space-2 allowed -> space-0 header skipped (no tabs), SpaceSkipped=1
```

## Steps

1. Same two-window fixture as keeps-allowed.
2. SpaceAllow = {2}.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	space0, space2 := 0, 2
	s0 := baseSession(1, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "on-0", "/dev/ttys001", "Default")
	s2 := baseSession(1, "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", "on-2", "/dev/ttys002", "Default")
	w0 := oneSessionWindow(1, "Desk0", 10, 1, "T1", s0)
	w0.FixedSpace = &space0
	w2 := oneSessionWindow(2, "Desk2", 20, 1, "T1", s2)
	w2.FixedSpace = &space2
	req.Windows = []snapshot.SnapshotWindow{w0, w2}
	req.IdleTTYs = []string{"ttys001", "ttys002"}
	req.SpaceAllow = []int{2}
	return nil
}
```
