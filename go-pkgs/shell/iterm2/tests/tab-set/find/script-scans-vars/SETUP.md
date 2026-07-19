# Scenario

**Feature**: find script scans both tab-set marker variables for a set name

```
BuildTabSetFindScript("bots")
  -> AppleScript mentions user.koolTabSet, user.koolTabSetTab, and set name bots
```

## Steps

1. Phase `build-find-script`.
2. Set name `bots`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "build-find-script"
	req.TabSetName = "bots"
	return nil
}
```
