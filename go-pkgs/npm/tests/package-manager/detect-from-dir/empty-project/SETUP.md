# Scenario

**Feature**: empty project directory yields unknown manager

```
# no indicators, no package.json
empty dir -> Manager unknown
```

## Steps

1. Create an empty temp project directory.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ProjectDir = writeProject(t, nil)
	return nil
}```
