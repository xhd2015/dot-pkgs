# Scenario

**Feature**: `--owner` passes explicit owner to `gh repo list`

```
# explicit owner flag
RunCLI repo list --owner alice -> gh repo list alice
```

## Steps

1. Mock auth and single-repo fixture for alice.
2. Set `req.Args` to `["repo", "list", "--owner", "alice"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"repo", "list", "--owner", "alice"}
	req.GhBin = writeOwnedOnlyGh(t, "alice", "testdata/repos.json")
	return nil
}
```