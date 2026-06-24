# Scenario

**Feature**: empty scan with `--json` returns empty array

```
empty workspace --json -> []
```

## Steps

1. Create empty workspace directory.
2. Set `req.Args` to `["--root", <workspace>, "--json"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	req.Args = []string{"--root", root, "--json"}
	return nil
}
```