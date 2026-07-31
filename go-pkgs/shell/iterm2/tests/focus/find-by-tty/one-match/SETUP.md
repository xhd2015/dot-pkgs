# Scenario

**Feature**: FindByTTY returns the single matching SessionRef

```
refs[ttys148, ttys149] + query [/dev/ttys149]
  -> FindByTTY -> [ref with ttys149]
```

## Steps

1. Two fixture refs; query matches the second only (same form as ref TTY).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Refs = []SessionRefInput{
		{WindowID: "win-1", TabIndex: 1, SessionID: "s1", TTY: "/dev/ttys148", Name: "A"},
		{WindowID: "win-2", TabIndex: 3, SessionID: "s2", TTY: "/dev/ttys149", Name: "B"},
	}
	req.QueryTTYs = []string{"/dev/ttys149"}
	return nil
}
```
