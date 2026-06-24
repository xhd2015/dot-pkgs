# Scenario

**Feature**: missing required `--root` flag

```
RunCLI (no --root) -> validation error on stderr
```

## Steps

1. Invoke with empty argv (no `--root`).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{}
	return nil
}
```