# Scenario

**Feature**: single-tab set uses the new window’s initial session only

```
TabSetSpec{Tabs×1}
  -> BuildTabSetNewWindowScript
  -> 1× create window, 0× create tab, one write text command
```

## Steps

1. Set name `solo`.
2. One tab with id `only`, command `solo-cmd`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.TabSetName = "solo"
	req.Tabs = []TabSpecInput{
		{ID: "only", Name: "Only", Command: "solo-cmd"},
	}
	return nil
}
```
