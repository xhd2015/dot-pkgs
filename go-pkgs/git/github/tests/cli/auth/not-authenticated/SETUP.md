# Scenario

**Feature**: unauthenticated `repo list` prints gh auth login hint on stderr

```
# auth gate failure
RunCLI repo list -> gh api user exit 4 -> stderr hint
```

## Steps

1. Mock `gh api user` auth failure.
2. Set `req.Args` to `["repo", "list"]` (infer owner path).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"repo", "list"}
	req.GhBin = writeAuthFailGh(t)
	return nil
}
```