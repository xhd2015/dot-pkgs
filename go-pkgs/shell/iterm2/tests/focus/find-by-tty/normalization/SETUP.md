# Scenario

**Feature**: FindByTTY matches when query and ref TTY forms differ

```
ref TTY=/dev/ttys148 + query ["ttys148"]
  -> FindByTTY -> [that ref]
```

## Steps

1. One ref with `/dev/ttys148`; query bare `ttys148`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Refs = []SessionRefInput{
		{WindowID: "win-norm", TabIndex: 1, SessionID: "s-norm", TTY: "/dev/ttys148", Name: "Norm"},
		{WindowID: "win-other", TabIndex: 1, SessionID: "s-other", TTY: "/dev/ttys200", Name: "Other"},
	}
	req.QueryTTYs = []string{"ttys148"}
	return nil
}
```
