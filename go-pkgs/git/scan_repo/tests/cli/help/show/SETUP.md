# Scenario

**Feature**: `--help` prints flag documentation

```
RunCLI --help -> stdout usage text
```

## Steps

1. Set `req.Args` to `["--help"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--help"}
	return nil
}
```