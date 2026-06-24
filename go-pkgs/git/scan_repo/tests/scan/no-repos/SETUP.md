# Scenario

**Feature**: empty workspace yields empty result

```
Walk finds no .git -> empty []Repo
```

## Steps

1. Create empty workspace directory.
2. Set `req.Roots` to the workspace.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	req.Roots = []string{root}
	return nil
}
```