# Scenario

**Feature**: NoSubmit=true stages command via write text without newline

```
TabSpec{Command: staged-cmd, NoSubmit: true}
  -> BuildTabSetNewWindowScript
  -> write text "staged-cmd" without newline
```

## Steps

1. One tab with command `staged-cmd` and NoSubmit=true.
2. Assert calls product API with NoSubmit set (Classic TDD: field may be absent → RED).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	req.TabSetName = "stage"
	req.Tabs = []TabSpecInput{
		{ID: "s1", Name: "Stage", Command: "staged-cmd", NoSubmit: true},
	}
	return nil
}
```
