# Scenario

**Feature**: unknown manager pref returns an error

```
# invalid pref
pref yarnberry -> error listing expected managers
```

## Steps

1. Create empty project directory.
2. Set `req.Pref` to `yarnberry`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ProjectDir = writeProject(t, nil)
	req.Pref = "yarnberry"
	return nil
}```
