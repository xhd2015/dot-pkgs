# Scenario

**Feature**: non-empty WindowName sets the new window’s name

```
TabSetSpec{WindowName=Bots Window, Tabs×1}
  -> BuildTabSetNewWindowScript
  -> set name of new window to "Bots Window"
```

## Steps

1. WindowName `Bots Window`.
2. One tab so script still creates a window.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {

	req.TabSetName = "bots"
	req.WindowName = "Bots Window"
	req.Tabs = []TabSpecInput{
		{ID: "w1", Name: "Main", Command: "echo-main"},
	}
	return nil
}
```
