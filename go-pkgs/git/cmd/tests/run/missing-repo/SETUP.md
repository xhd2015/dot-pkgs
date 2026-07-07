# Scenario

**Feature**: Run in a non-git directory returns error

```
plain temp dir (no .git) -> Run(rev-parse HEAD) -> error
```

## Steps

1. Use empty temp dir without git init.
2. Run `rev-parse HEAD`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Dir = t.TempDir()
	req.Args = []string{"rev-parse", "HEAD"}
	return nil
}
```
