# Scenario

**Feature**: FindByTTY returns empty when no TTY matches

```
refs[ttys148, ttys149] + query [ttys999]
  -> FindByTTY -> []
```

## Steps

1. Two fixture refs; query a TTY present in neither.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Refs = []SessionRefInput{
		{WindowID: "win-1", TabIndex: 1, SessionID: "s1", TTY: "/dev/ttys148", Name: "A"},
		{WindowID: "win-1", TabIndex: 2, SessionID: "s2", TTY: "/dev/ttys149", Name: "B"},
	}
	req.QueryTTYs = []string{"ttys999"}
	return nil
}
```
