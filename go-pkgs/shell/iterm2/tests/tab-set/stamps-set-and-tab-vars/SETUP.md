# Scenario

**Feature**: every tab stamps set id and per-tab id session variables

```
TabSetSpec{Name=bots, Tabs with IDs t1,t2}
  -> BuildTabSetNewWindowScript
  -> each session: variable user.koolTabSet="bots", user.koolTabSetTab=<tab id>
```

## Steps

1. Set name `bots`.
2. Two tabs with IDs `t1` and `t2`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.TabSetName = "bots"
	req.Tabs = []TabSpecInput{
		{ID: "t1", Name: "One", Command: "echo-one"},
		{ID: "t2", Name: "Two", Command: "echo-two"},
	}
	return nil
}
```
