# Scenario

**Feature**: `--json` with empty result prints `[]`

```
# no repos
RunCLI repo list --owner alice --json -> empty JSON array
```

## Steps

1. Mock auth and empty `gh repo list alice` response.
2. Set `req.Args` with `--json`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"repo", "list", "--owner", "alice", "--json"}
	req.GhBin = writeEmptyOwnedGh(t, "alice")
	return nil
}
```