# Scenario

**Feature**: trailing positional args after `repo list` flags are rejected

```
# extra positional token
RunCLI repo list extra -> unexpected arguments -> stderr
```

## Steps

1. Set `req.Args` to `["repo", "list", "extra"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"repo", "list", "extra"}
	return nil
}
```