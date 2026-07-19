# Scenario

**Feature**: non-empty TabSpec.Name becomes the session name in AppleScript

```
TabSetSpec{Tabs with Name=Alpha, Beta}
  -> BuildTabSetNewWindowScript
  -> set name of session (or equivalent) to Alpha / Beta
```

## Steps

1. Two tabs with distinct display names `Alpha` and `Beta`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.TabSetName = "named"
	req.Tabs = []TabSpecInput{
		{ID: "n1", Name: "Alpha", Command: "run-alpha"},
		{ID: "n2", Name: "Beta", Command: "run-beta"},
	}
	return nil
}
```
