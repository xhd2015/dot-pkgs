# Scenario

**Feature**: `--json` prints indented `RepoResult` array

```
# multi-repo JSON output
RunCLI repo list --owner alice --json -> JSON array with matched_by
```

## Steps

1. Mock auth and `gh repo list alice` with two-repo fixture.
2. Set `req.Args` with `--json`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"repo", "list", "--owner", "alice", "--json"}
	req.GhBin = writeOwnedOnlyGh(t, "alice", "testdata/repos.json")
	return nil
}
```