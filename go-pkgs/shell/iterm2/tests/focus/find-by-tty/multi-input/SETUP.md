# Scenario

**Feature**: multi query TTYs yield union of matches in stable ref order

```
refs order: A(ttys001), B(ttys002), C(ttys003)
query: [ttys003, ttys001]   # query order ≠ ref order
  -> FindByTTY -> [A, C]    # preserve refs order, not query order
```

## Steps

1. Three refs; two query TTYs that hit first and third.
2. Assert result order follows `refs`, not `QueryTTYs`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Refs = []SessionRefInput{
		{WindowID: "wA", TabIndex: 1, SessionID: "sA", TTY: "/dev/ttys001", Name: "A"},
		{WindowID: "wB", TabIndex: 1, SessionID: "sB", TTY: "/dev/ttys002", Name: "B"},
		{WindowID: "wC", TabIndex: 2, SessionID: "sC", TTY: "/dev/ttys003", Name: "C"},
	}
	// Query order intentionally reversed vs appearance in refs.
	req.QueryTTYs = []string{"ttys003", "ttys001"}
	return nil
}
```
