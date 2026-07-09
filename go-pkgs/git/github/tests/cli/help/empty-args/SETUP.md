# Scenario

**Feature**: empty `RunCLI` args print top-level usage (same class as `--help`)

```
# empty argv after kool github
RunCLI [] -> top-level usage mentioning repo -> stdout, exit 0
```

## Steps

1. Set `req.Args` to `[]string{}` (empty).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{}
	return nil
}
```
