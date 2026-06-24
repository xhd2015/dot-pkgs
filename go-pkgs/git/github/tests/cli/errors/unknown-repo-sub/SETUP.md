# Scenario

**Feature**: unrecognized `repo` subcommand

```
# repo with unknown verb
RunCLI repo nope -> unrecognized repo command -> stderr
```

## Steps

1. Set `req.Args` to `["repo", "nope"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"repo", "nope"}
	return nil
}
```